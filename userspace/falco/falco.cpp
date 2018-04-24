/*
Copyright (C) 2016 Draios inc.

This file is part of falco.

falco is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License version 2 as
published by the Free Software Foundation.

falco is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with falco.  If not, see <http://www.gnu.org/licenses/>.
*/

#define __STDC_FORMAT_MACROS

#include <stdio.h>
#include <fstream>
#include <set>
#include <list>
#include <string>
#include <signal.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <unistd.h>
#include <getopt.h>

#include <sinsp.h>

#include "logger.h"

#include "configuration.h"
#include "falco_engine.h"
#include "config_falco.h"
#include "statsfilewriter.h"

bool g_terminate = false;
bool g_reopen_outputs = false;

//
// Helper functions
//
static void signal_callback(int signal)
{
	g_terminate = true;
}

static void reopen_outputs(int signal)
{
	g_reopen_outputs = true;
}

//
// Program help
//
static void usage()
{
    printf(
	   "falco version " FALCO_VERSION "\n"
	   "Usage: falco [options]\n\n"
	   "Options:\n"
	   " -h, --help                    Print this page\n"
	   " -c                            Configuration file (default " FALCO_SOURCE_CONF_FILE ", " FALCO_INSTALL_CONF_FILE ")\n"
	   " -A                            Monitor all events, including those with EF_DROP_FALCO flag.\n"
	   " -d, --daemon                  Run as a daemon\n"
	   " -D <pattern>                  Disable any rules matching the regex <pattern>. Can be specified multiple times.\n"
	   "                               Can not be specified with -t.\n"
           " -e <events_file>              Read the events from <events_file> (in .scap format) instead of tapping into live.\n"
	   " -k <url>, --k8s-api=<url>\n"
	   "                               Enable Kubernetes support by connecting to the API server\n"
      	   "                               specified as argument. E.g. \"http://admin:password@127.0.0.1:8080\".\n"
	   "                               The API server can also be specified via the environment variable\n"
	   "                               FALCO_K8S_API.\n"
	   " -K <bt_file> | <cert_file>:<key_file[#password]>[:<ca_cert_file>], --k8s-api-cert=<bt_file> | <cert_file>:<key_file[#password]>[:<ca_cert_file>]\n"
	   "                               Use the provided files names to authenticate user and (optionally) verify the K8S API\n"
	   "                               server identity.\n"
	   "                               Each entry must specify full (absolute, or relative to the current directory) path\n"
	   "                               to the respective file.\n"
	   "                               Private key password is optional (needed only if key is password protected).\n"
	   "                               CA certificate is optional. For all files, only PEM file format is supported. \n"
	   "                               Specifying CA certificate only is obsoleted - when single entry is provided \n"
	   "                               for this option, it will be interpreted as the name of a file containing bearer token.\n"
	   "                               Note that the format of this command-line option prohibits use of files whose names contain\n"
	   "                               ':' or '#' characters in the file name.\n"
	   " -L                            Show the name and description of all rules and exit.\n"
	   " -l <rule>                     Show the name and description of the rule with name <rule> and exit.\n"
	   " -m <url[,marathon_url]>, --mesos-api=<url[,marathon_url]>\n"
	   "                               Enable Mesos support by connecting to the API server\n"
	   "                               specified as argument. E.g. \"http://admin:password@127.0.0.1:5050\".\n"
	   "                               Marathon url is optional and defaults to Mesos address, port 8080.\n"
	   "                               The API servers can also be specified via the environment variable\n"
	   "                               FALCO_MESOS_API.\n"
	   " -M <num_seconds>              Stop collecting after <num_seconds> reached.\n"
	   " -o, --option <key>=<val>      Set the value of option <key> to <val>. Overrides values in configuration file.\n"
	   "                               <key> can be a two-part <key>.<subkey>\n"
	   " -p <output_format>, --print=<output_format>\n"
	   "                               Add additional information to each falco notification's output.\n"
	   "                               With -pc or -pcontainer will use a container-friendly format.\n"
	   "                               With -pk or -pkubernetes will use a kubernetes-friendly format.\n"
	   "                               With -pm or -pmesos will use a mesos-friendly format.\n"
	   "                               Additionally, specifying -pc/-pk/-pm will change the interpretation\n"
	   "                               of %%container.info in rule output fields\n"
	   "                               See the examples section below for more info.\n"
	   " -P, --pidfile <pid_file>      When run as a daemon, write pid to specified file\n"
           " -r <rules_file>               Rules file/directory (defaults to value set in configuration file,\n"
           "                               or /etc/falco_rules.yaml). Can be specified multiple times to read\n"
           "                               from multiple files/directories.\n"
	   " -s <stats_file>               If specified, write statistics related to falco's reading/processing of events\n"
	   "                               to this file. (Only useful in live mode).\n"
	   " -T <tag>                      Disable any rules with a tag=<tag>. Can be specified multiple times.\n"
	   "                               Can not be specified with -t.\n"
	   " -t <tag>                      Only run those rules with a tag=<tag>. Can be specified multiple times.\n"
	   "                               Can not be specified with -T/-D.\n"
	   " -U,--unbuffered               Turn off output buffering to configured outputs. This causes every\n"
	   "                               single line emitted by falco to be flushed, which generates higher CPU\n"
	   "                               usage but is useful when piping those outputs into another process\n"
	   "                               or into a script.\n"
	   " -V,--validate <rules_file>    Read the contents of the specified rules(s) file and exit\n"
	   "                               Can be specified multiple times to validate multiple files.\n"
	   " -v                            Verbose output.\n"
           " --version                     Print version number.\n"
	   "\n"
    );
}

static void display_fatal_err(const string &msg)
{
	falco_logger::log(LOG_ERR, msg);

	/**
	 * If stderr logging is not enabled, also log to stderr. When
	 * daemonized this will simply write to /dev/null.
	 */
	if (! falco_logger::log_stderr)
	{
		std::cerr << msg;
	}
}

// Splitting into key=value or key.subkey=value will be handled by configuration class.
std::list<string> cmdline_options;

//
// Event processing loop
//
uint64_t do_inspect(falco_engine *engine,
		    falco_outputs *outputs,
		    sinsp* inspector,
		    uint64_t duration_to_tot_ns,
		    string &stats_filename,
		    bool all_events)
{
	uint64_t num_evts = 0;
	int32_t res;
	sinsp_evt* ev;
	StatsFileWriter writer;
	uint64_t duration_start = 0;

	if (stats_filename != "")
	{
		string errstr;

		if (!writer.init(inspector, stats_filename, 5, errstr))
		{
			throw falco_exception(errstr);
		}
	}

	//
	// Loop through the events
	//
	while(1)
	{

		res = inspector->next(&ev);

		writer.handle();

		if(g_reopen_outputs)
		{
			outputs->reopen_outputs();
			g_reopen_outputs = false;
		}

		if (g_terminate)
		{
			break;
		}
		else if(res == SCAP_TIMEOUT)
		{
			continue;
		}
		else if(res == SCAP_EOF)
		{
			break;
		}
		else if(res != SCAP_SUCCESS)
		{
			//
			// Event read error.
			// Notify the chisels that we're exiting, and then die with an error.
			//
			cerr << "res = " << res << endl;
			throw sinsp_exception(inspector->getlasterr().c_str());
		}

		if (duration_start == 0)
		{
			duration_start = ev->get_ts();
		} else if(duration_to_tot_ns > 0)
		{
			if(ev->get_ts() - duration_start >= duration_to_tot_ns)
			{
				break;
			}
		}

		if(!ev->falco_consider() && !all_events)
		{
			continue;
		}

		// As the inspector has no filter at its level, all
		// events are returned here. Pass them to the falco
		// engine, which will match the event against the set
		// of rules. If a match is found, pass the event to
		// the outputs.
		unique_ptr<falco_engine::rule_result> res = engine->process_event(ev);
		if(res)
		{
			outputs->handle_event(res->evt, res->rule, res->priority_num, res->format);
		}

		num_evts++;
	}

	return num_evts;
}

//
// ARGUMENT PARSING AND PROGRAM SETUP
//
int falco_init(int argc, char **argv)
{
	int result = EXIT_SUCCESS;
	sinsp* inspector = NULL;
	falco_engine *engine = NULL;
	falco_outputs *outputs = NULL;
	int op;
	int long_index = 0;
	string scap_filename;
	string conf_filename;
	string outfile;
	list<string> rules_filenames;
	bool daemon = false;
	string pidfilename = "/var/run/falco.pid";
	bool describe_all_rules = false;
	string describe_rule = "";
	list<string> validate_rules_filenames;
	string stats_filename = "";
	bool verbose = false;
	bool all_events = false;
	string* k8s_api = 0;
	string* k8s_api_cert = 0;
	string* mesos_api = 0;
	string output_format = "";
	bool replace_container_info = false;
	int duration_to_tot = 0;

	// Used for writing trace files
	int duration_seconds = 0;
	int rollover_mb = 0;
	int file_limit = 0;
	unsigned long event_limit = 0L;
	bool compress = false;
	bool buffered_outputs = true;
	bool buffered_cmdline = false;

	// Used for stats
	uint64_t num_evts;
	double duration;
	scap_stats cstats;

	static struct option long_options[] =
	{
		{"help", no_argument, 0, 'h' },
		{"daemon", no_argument, 0, 'd' },
		{"k8s-api", required_argument, 0, 'k'},
		{"k8s-api-cert", required_argument, 0, 'K' },
		{"mesos-api", required_argument, 0, 'm'},
		{"option", required_argument, 0, 'o'},
		{"print", required_argument, 0, 'p' },
		{"pidfile", required_argument, 0, 'P' },
		{"unbuffered", no_argument, 0, 'U' },
		{"version", no_argument, 0, 0 },
		{"validate", required_argument, 0, 'V' },
		{"writefile", required_argument, 0, 'w' },

		{0, 0, 0, 0}
	};

	try
	{
		set<string> disabled_rule_patterns;
		string pattern;
		string all_rules = ".*";
		set<string> disabled_rule_tags;
		set<string> enabled_rule_tags;

		//
		// Parse the args
		//
		while((op = getopt_long(argc, argv,
                                        "hc:AdD:e:k:K:Ll:m:M:o:P:p:r:s:T:t:UvV:w:",
                                        long_options, &long_index)) != -1)
		{
			switch(op)
			{
			case 'h':
				usage();
				goto exit;
			case 'c':
				conf_filename = optarg;
				break;
			case 'A':
				all_events = true;
				break;
			case 'd':
				daemon = true;
				break;
			case 'D':
				pattern = optarg;
				disabled_rule_patterns.insert(pattern);
				break;
			case 'e':
				scap_filename = optarg;
				k8s_api = new string();
				mesos_api = new string();
				break;
			case 'k':
				k8s_api = new string(optarg);
				break;
			case 'K':
				k8s_api_cert = new string(optarg);
				break;
			case 'L':
				describe_all_rules = true;
				break;
			case 'l':
				describe_rule = optarg;
				break;
			case 'm':
				mesos_api = new string(optarg);
				break;
			case 'M':
				duration_to_tot = atoi(optarg);
				if(duration_to_tot <= 0)
				{
					throw sinsp_exception(string("invalid duration") + optarg);
				}
				break;
			case 'o':
				cmdline_options.push_back(optarg);
				break;
			case 'P':
				pidfilename = optarg;
				break;
			case 'p':
				if(string(optarg) == "c" || string(optarg) == "container")
				{
					output_format = "container=%container.name (id=%container.id)";
					replace_container_info = true;
				}
				else if(string(optarg) == "k" || string(optarg) == "kubernetes")
				{
					output_format = "k8s.pod=%k8s.pod.name container=%container.id";
					replace_container_info = true;
				}
				else if(string(optarg) == "m" || string(optarg) == "mesos")
				{
					output_format = "task=%mesos.task.name container=%container.id";
					replace_container_info = true;
				}
				else
				{
					output_format = optarg;
					replace_container_info = false;
				}
				break;
			case 'r':
				falco_configuration::read_rules_file_directory(string(optarg), rules_filenames);
				break;
			case 's':
				stats_filename = optarg;
				break;
			case 'T':
				disabled_rule_tags.insert(optarg);
				break;
			case 't':
				enabled_rule_tags.insert(optarg);
				break;
			case 'U':
				buffered_outputs = false;
				buffered_cmdline = true;
				break;
			case 'v':
				verbose = true;
				break;
			case 'V':
				validate_rules_filenames.push_back(optarg);
				break;
			case 'w':
				outfile = optarg;
				break;
			case '?':
				result = EXIT_FAILURE;
				goto exit;
			default:
				break;
			}

		}

		if(string(long_options[long_index].name) == "version")
		{
			printf("falco version %s\n", FALCO_VERSION);
			return EXIT_SUCCESS;
		}


		inspector = new sinsp();
		engine = new falco_engine();
		engine->set_inspector(inspector);
		engine->set_extra(output_format, replace_container_info);

		outputs = new falco_outputs();
		outputs->set_inspector(inspector);

		// Some combinations of arguments are not allowed.
		if (daemon && pidfilename == "") {
			throw std::invalid_argument("If -d is provided, a pid file must also be provided");
		}

		ifstream conf_stream;
		if (conf_filename.size())
		{
			conf_stream.open(conf_filename);
			if (!conf_stream.is_open())
			{
				throw std::runtime_error("Could not find configuration file at " + conf_filename);
			}
		}
		else
		{
			conf_stream.open(FALCO_SOURCE_CONF_FILE);
			if (conf_stream.is_open())
			{
				conf_filename = FALCO_SOURCE_CONF_FILE;
			}
			else
			{
				conf_stream.open(FALCO_INSTALL_CONF_FILE);
				if (conf_stream.is_open())
				{
					conf_filename = FALCO_INSTALL_CONF_FILE;
				}
				else
				{
					conf_filename = "";
				}
			}
		}

		if(validate_rules_filenames.size() > 0)
		{
			falco_logger::log(LOG_INFO, "Validating rules file(s):\n");
			for(auto file : validate_rules_filenames)
			{
				falco_logger::log(LOG_INFO, "   " + file + "\n");
			}
			for(auto file : validate_rules_filenames)
			{
				engine->load_rules_file(file, verbose, all_events);
			}
			falco_logger::log(LOG_INFO, "Ok\n");
			goto exit;
		}

		falco_configuration config;
		if (conf_filename.size())
		{
			config.init(conf_filename, cmdline_options);
			// log after config init because config determines where logs go
			falco_logger::log(LOG_INFO, "Falco initialized with configuration file " + conf_filename + "\n");
		}
		else
		{
			config.init(cmdline_options);
			falco_logger::log(LOG_INFO, "Falco initialized. No configuration file found, proceeding with defaults\n");
		}

		if (rules_filenames.size())
		{
			config.m_rules_filenames = rules_filenames;
		}

		engine->set_min_priority(config.m_min_priority);

		if(buffered_cmdline)
		{
			config.m_buffered_outputs = buffered_outputs;
		}

		if(config.m_rules_filenames.size() == 0)
		{
			throw std::invalid_argument("You must specify at least one rules file/directory via -r or a rules_file entry in falco.yaml");
		}

		falco_logger::log(LOG_DEBUG, "Configured rules filenames:\n");
		for (auto filename : config.m_rules_filenames)
		{
			falco_logger::log(LOG_DEBUG, string("   ") + filename + "\n");
		}

		for (auto filename : config.m_rules_filenames)
		{
			falco_logger::log(LOG_INFO, "Loading rules from file " + filename + ":\n");
			engine->load_rules_file(filename, verbose, all_events);
		}

		// You can't both disable and enable rules
		if((disabled_rule_patterns.size() + disabled_rule_tags.size() > 0) &&
		    enabled_rule_tags.size() > 0) {
			throw std::invalid_argument("You can not specify both disabled (-D/-T) and enabled (-t) rules");
		}

		for (auto pattern : disabled_rule_patterns)
		{
			falco_logger::log(LOG_INFO, "Disabling rules matching pattern: " + pattern + "\n");
			engine->enable_rule(pattern, false);
		}

		if(disabled_rule_tags.size() > 0)
		{
			for(auto tag : disabled_rule_tags)
			{
				falco_logger::log(LOG_INFO, "Disabling rules with tag: " + tag + "\n");
			}
			engine->enable_rule_by_tag(disabled_rule_tags, false);
		}

		if(enabled_rule_tags.size() > 0)
		{

			// Since we only want to enable specific
			// rules, first disable all rules.
			engine->enable_rule(all_rules, false);
			for(auto tag : enabled_rule_tags)
			{
				falco_logger::log(LOG_INFO, "Enabling rules with tag: " + tag + "\n");
			}
			engine->enable_rule_by_tag(enabled_rule_tags, true);
		}

		outputs->init(config.m_json_output,
			      config.m_json_include_output_property,
			      config.m_notifications_rate, config.m_notifications_max_burst,
			      config.m_buffered_outputs);

		if(!all_events)
		{
			inspector->set_drop_event_flags(EF_DROP_FALCO);
			inspector->start_dropping_mode(1);
		}

		if (describe_all_rules)
		{
			engine->describe_rule(NULL);
			goto exit;
		}

		if (describe_rule != "")
		{
			engine->describe_rule(&describe_rule);
			goto exit;
		}

		inspector->set_hostname_and_port_resolution_mode(false);

		for(auto output : config.m_outputs)
		{
			outputs->add_output(output);
		}

		if(signal(SIGINT, signal_callback) == SIG_ERR)
		{
			fprintf(stderr, "An error occurred while setting SIGINT signal handler.\n");
			result = EXIT_FAILURE;
			goto exit;
		}

		if(signal(SIGTERM, signal_callback) == SIG_ERR)
		{
			fprintf(stderr, "An error occurred while setting SIGTERM signal handler.\n");
			result = EXIT_FAILURE;
			goto exit;
		}

		if(signal(SIGUSR1, reopen_outputs) == SIG_ERR)
		{
			fprintf(stderr, "An error occurred while setting SIGUSR1 signal handler.\n");
			result = EXIT_FAILURE;
			goto exit;
		}

		if (scap_filename.size())
		{
			inspector->open(scap_filename);
		}
		else
		{
			try
			{
				inspector->open(200);
			}
			catch(sinsp_exception e)
			{
				if(system("modprobe " PROBE_NAME " > /dev/null 2> /dev/null"))
				{
					falco_logger::log(LOG_ERR, "Unable to load the driver. Exiting.\n");
				}
				inspector->open();
			}
		}

		// If daemonizing, do it here so any init errors will
		// be returned in the foreground process.
		if (daemon) {
			pid_t pid, sid;

			pid = fork();
			if (pid < 0) {
				// error
				falco_logger::log(LOG_ERR, "Could not fork. Exiting.\n");
				result = EXIT_FAILURE;
				goto exit;
			} else if (pid > 0) {
				// parent. Write child pid to pidfile and exit
				std::ofstream pidfile;
				pidfile.open(pidfilename);

				if (!pidfile.good())
				{
					falco_logger::log(LOG_ERR, "Could not write pid to pid file " + pidfilename + ". Exiting.\n");
					result = EXIT_FAILURE;
					goto exit;
				}
				pidfile << pid;
				pidfile.close();
				goto exit;
			}
			// if here, child.

			// Become own process group.
			sid = setsid();
			if (sid < 0) {
				falco_logger::log(LOG_ERR, "Could not set session id. Exiting.\n");
				result = EXIT_FAILURE;
				goto exit;
			}

			// Set umask so no files are world anything or group writable.
			umask(027);

			// Change working directory to '/'
			if ((chdir("/")) < 0) {
				falco_logger::log(LOG_ERR, "Could not change working directory to '/'. Exiting.\n");
				result = EXIT_FAILURE;
				goto exit;
			}

			// Close stdin, stdout, stderr and reopen to /dev/null
			close(0);
			close(1);
			close(2);
			open("/dev/null", O_RDONLY);
			open("/dev/null", O_RDWR);
			open("/dev/null", O_RDWR);
		}

		if(outfile != "")
		{
			inspector->setup_cycle_writer(outfile, rollover_mb, duration_seconds, file_limit, event_limit, compress);
			inspector->autodump_next_file();
		}

		duration = ((double)clock()) / CLOCKS_PER_SEC;

		//
		// run k8s, if required
		//
		if(k8s_api)
		{
			if(!k8s_api_cert)
			{
				if(char* k8s_cert_env = getenv("FALCO_K8S_API_CERT"))
				{
					k8s_api_cert = new string(k8s_cert_env);
				}
			}
			inspector->init_k8s_client(k8s_api, k8s_api_cert, verbose);
			k8s_api = 0;
			k8s_api_cert = 0;
		}
		else if(char* k8s_api_env = getenv("FALCO_K8S_API"))
		{
			if(k8s_api_env != NULL)
			{
				if(!k8s_api_cert)
				{
					if(char* k8s_cert_env = getenv("FALCO_K8S_API_CERT"))
					{
						k8s_api_cert = new string(k8s_cert_env);
					}
				}
				k8s_api = new string(k8s_api_env);
				inspector->init_k8s_client(k8s_api, k8s_api_cert, verbose);
			}
			else
			{
				delete k8s_api;
				delete k8s_api_cert;
			}
			k8s_api = 0;
			k8s_api_cert = 0;
		}

		//
		// run mesos, if required
		//
		if(mesos_api)
		{
			inspector->init_mesos_client(mesos_api, verbose);
		}
		else if(char* mesos_api_env = getenv("FALCO_MESOS_API"))
		{
			if(mesos_api_env != NULL)
			{
				mesos_api = new string(mesos_api_env);
				inspector->init_mesos_client(mesos_api, verbose);
			}
		}
		delete mesos_api;
		mesos_api = 0;

		num_evts = do_inspect(engine,
				      outputs,
				      inspector,
				      uint64_t(duration_to_tot*ONE_SECOND_IN_NS),
				      stats_filename,
				      all_events);

		duration = ((double)clock()) / CLOCKS_PER_SEC - duration;

		inspector->get_capture_stats(&cstats);

		if(verbose)
		{
			fprintf(stderr, "Driver Events:%" PRIu64 "\nDriver Drops:%" PRIu64 "\n",
				cstats.n_evts,
				cstats.n_drops);

			fprintf(stderr, "Elapsed time: %.3lf, Captured Events: %" PRIu64 ", %.2lf eps\n",
				duration,
				num_evts,
				num_evts / duration);
		}

		inspector->close();

		engine->print_stats();
	}
	catch(exception &e)
	{
		display_fatal_err("Runtime error: " + string(e.what()) + ". Exiting.\n");

		result = EXIT_FAILURE;
	}

exit:

	delete inspector;
	delete engine;
	delete outputs;

	return result;
}

//
// MAIN
//
int main(int argc, char **argv)
{
	return falco_init(argc, argv);
}
