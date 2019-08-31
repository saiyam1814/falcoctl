package kubernetesfalc

var defaultFalcoConfig = map[string]string{
	"falco_rules.local.yanl": `#
# Copyright (C) 2016-2018 Draios Inc dba Sysdig.
#
# This file is part of falco.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

####################
# Your custom rules!
####################

# Add new rules, like this one
# - rule: The program "sudo" is run in a container
#   desc: An event will trigger every time you run sudo in a container
#   condition: evt.type = execve and evt.dir=< and container.id != host and proc.name = sudo
#   output: "Sudo run in container (user=%user.name %container.info parent=%proc.pname cmdline=%proc.cmdline)"
#   priority: ERROR
#   tags: [users, container]

# Or override/append to any rule, macro, or list from the Default Rules
`,
	"falco_rules.yaml": `#
# Copyright (C) 2016-2018 Draios Inc dba Sysdig.
#
# This file is part of falco.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# See xxx for details on falco engine and rules versioning. Currently,
# this specific rules file is compatible with engine version 0
# (e.g. falco releases <= 0.13.1), so we'll keep the
# required_engine_version lines commented out, so maintain
# compatibility with older falco releases. With the first incompatible
# change to this rules file, we'll uncomment this line and set it to
# the falco engine version in use at the time.
#
#- required_engine_version: 2

# Currently disabled as read/write are ignored syscalls. The nearly
# similar open_write/open_read check for files being opened for
# reading/writing.
# - macro: write
#   condition: (syscall.type=write and fd.type in (file, directory))
# - macro: read
#   condition: (syscall.type=read and evt.dir=> and fd.type in (file, directory))

- macro: open_write
  condition: (evt.type=open or evt.type=openat) and evt.is_open_write=true and fd.typechar='f' and fd.num>=0

- macro: open_read
  condition: (evt.type=open or evt.type=openat) and evt.is_open_read=true and fd.typechar='f' and fd.num>=0

- macro: open_directory
  condition: (evt.type=open or evt.type=openat) and evt.is_open_read=true and fd.typechar='d' and fd.num>=0

- macro: never_true
  condition: (evt.num=0)

- macro: always_true
  condition: (evt.num>=0)

# In some cases, such as dropped system call events, information about
# the process name may be missing. For some rules that really depend
# on the identity of the process performing an action such as opening
# a file, etc., we require that the process name be known.
- macro: proc_name_exists
  condition: (proc.name!="<NA>")

- macro: rename
  condition: evt.type in (rename, renameat)
- macro: mkdir
  condition: evt.type in (mkdir, mkdirat)
- macro: remove
  condition: evt.type in (rmdir, unlink, unlinkat)

- macro: modify
  condition: rename or remove

- macro: spawned_process
  condition: evt.type = execve and evt.dir=<

- macro: create_symlink
  condition: evt.type in (symlink, symlinkat) and evt.dir=<

- macro: chmod
  condition: (evt.type in (chmod, fchmod, fchmodat) and evt.dir=<)

# File categories
- macro: bin_dir
  condition: fd.directory in (/bin, /sbin, /usr/bin, /usr/sbin)

- macro: bin_dir_mkdir
  condition: >
    (evt.arg[1] startswith /bin/ or
     evt.arg[1] startswith /sbin/ or
     evt.arg[1] startswith /usr/bin/ or
     evt.arg[1] startswith /usr/sbin/)

- macro: bin_dir_rename
  condition: >
    evt.arg[1] startswith /bin/ or
    evt.arg[1] startswith /sbin/ or
    evt.arg[1] startswith /usr/bin/ or
    evt.arg[1] startswith /usr/sbin/

- macro: etc_dir
  condition: fd.name startswith /etc/

# This detects writes immediately below / or any write anywhere below /root
- macro: root_dir
  condition: ((fd.directory=/ or fd.name startswith /root) and fd.name contains "/")

- list: shell_binaries
  items: [ash, bash, csh, ksh, sh, tcsh, zsh, dash]

- list: ssh_binaries
  items: [
    sshd, sftp-server, ssh-agent,
    ssh, scp, sftp,
    ssh-keygen, ssh-keysign, ssh-keyscan, ssh-add
    ]

- list: shell_mgmt_binaries
  items: [add-shell, remove-shell]

- macro: shell_procs
  condition: proc.name in (shell_binaries)

- list: coreutils_binaries
  items: [
    truncate, sha1sum, numfmt, fmt, fold, uniq, cut, who,
    groups, csplit, sort, expand, printf, printenv, unlink, tee, chcon, stat,
    basename, split, nice, "yes", whoami, sha224sum, hostid, users, stdbuf,
    base64, unexpand, cksum, od, paste, nproc, pathchk, sha256sum, wc, test,
    comm, arch, du, factor, sha512sum, md5sum, tr, runcon, env, dirname,
    tsort, join, shuf, install, logname, pinky, nohup, expr, pr, tty, timeout,
    tail, "[", seq, sha384sum, nl, head, id, mkfifo, sum, dircolors, ptx, shred,
    tac, link, chroot, vdir, chown, touch, ls, dd, uname, "true", pwd, date,
    chgrp, chmod, mktemp, cat, mknod, sync, ln, "false", rm, mv, cp, echo,
    readlink, sleep, stty, mkdir, df, dir, rmdir, touch
    ]

# dpkg -L login | grep bin | xargs ls -ld | grep -v '^d' | awk '{print $9}' | xargs -L 1 basename | tr "\\n" ","
- list: login_binaries
  items: [
    login, systemd, '"(systemd)"', systemd-logind, su,
    nologin, faillog, lastlog, newgrp, sg
    ]

# dpkg -L passwd | grep bin | xargs ls -ld | grep -v '^d' | awk '{print $9}' | xargs -L 1 basename | tr "\\n" ","
- list: passwd_binaries
  items: [
    shadowconfig, grpck, pwunconv, grpconv, pwck,
    groupmod, vipw, pwconv, useradd, newusers, cppw, chpasswd, usermod,
    groupadd, groupdel, grpunconv, chgpasswd, userdel, chage, chsh,
    gpasswd, chfn, expiry, passwd, vigr, cpgr, adduser, addgroup, deluser, delgroup
    ]

# repoquery -l shadow-utils | grep bin | xargs ls -ld | grep -v '^d' |
#     awk '{print $9}' | xargs -L 1 basename | tr "\\n" ","
- list: shadowutils_binaries
  items: [
    chage, gpasswd, lastlog, newgrp, sg, adduser, deluser, chpasswd,
    groupadd, groupdel, addgroup, delgroup, groupmems, groupmod, grpck, grpconv, grpunconv,
    newusers, pwck, pwconv, pwunconv, useradd, userdel, usermod, vigr, vipw, unix_chkpwd
    ]

- list: sysdigcloud_binaries
  items: [setup-backend, dragent, sdchecks]

- list: docker_binaries
  items: [docker, dockerd, exe, docker-compose, docker-entrypoi, docker-runc-cur, docker-current, dockerd-current]

- list: k8s_binaries
  items: [hyperkube, skydns, kube2sky, exechealthz, weave-net, loopback, bridge, openshift-sdn, openshift]

- list: lxd_binaries
  items: [lxd, lxcfs]

- list: http_server_binaries
  items: [nginx, httpd, httpd-foregroun, lighttpd, apache, apache2]

- list: db_server_binaries
  items: [mysqld, postgres, sqlplus]

- list: mysql_mgmt_binaries
  items: [mysql_install_d, mysql_ssl_rsa_s]

- list: postgres_mgmt_binaries
  items: [pg_dumpall, pg_ctl, pg_lsclusters, pg_ctlcluster]

- list: db_mgmt_binaries
  items: [mysql_mgmt_binaries, postgres_mgmt_binaries]

- list: nosql_server_binaries
  items: [couchdb, memcached, redis-server, rabbitmq-server, mongod]

- list: gitlab_binaries
  items: [gitlab-shell, gitlab-mon, gitlab-runner-b, git]

- list: interpreted_binaries
  items: [lua, node, perl, perl5, perl6, php, python, python2, python3, ruby, tcl]

- macro: interpreted_procs
  condition: >
    (proc.name in (interpreted_binaries))

- macro: server_procs
  condition: proc.name in (http_server_binaries, db_server_binaries, docker_binaries, sshd)

# The explicit quotes are needed to avoid the - characters being
# interpreted by the filter expression.
- list: rpm_binaries
  items: [dnf, rpm, rpmkey, yum, '"75-system-updat"', rhsmcertd-worke, subscription-ma,
          repoquery, rpmkeys, rpmq, yum-cron, yum-config-mana, yum-debug-dump,
          abrt-action-sav, rpmdb_stat, microdnf, rhn_check, yumdb]

- list: openscap_rpm_binaries
  items: [probe_rpminfo, probe_rpmverify, probe_rpmverifyfile, probe_rpmverifypackage]

- macro: rpm_procs
  condition: (proc.name in (rpm_binaries, openscap_rpm_binaries) or proc.name in (salt-minion))

- list: deb_binaries
  items: [dpkg, dpkg-preconfigu, dpkg-reconfigur, dpkg-divert, apt, apt-get, aptitude,
    frontend, preinst, add-apt-reposit, apt-auto-remova, apt-key,
    apt-listchanges, unattended-upgr, apt-add-reposit, apt-config, apt-cache
    ]

# The truncated dpkg-preconfigu is intentional, process names are
# truncated at the sysdig level.
- list: package_mgmt_binaries
  items: [rpm_binaries, deb_binaries, update-alternat, gem, pip, pip3, sane-utils.post, alternatives, chef-client, apk]

- macro: package_mgmt_procs
  condition: proc.name in (package_mgmt_binaries)

- macro: package_mgmt_ancestor_procs
  condition: proc.pname in (package_mgmt_binaries) or
             proc.aname[2] in (package_mgmt_binaries) or
             proc.aname[3] in (package_mgmt_binaries) or
             proc.aname[4] in (package_mgmt_binaries)

- macro: coreos_write_ssh_dir
  condition: (proc.name=update-ssh-keys and fd.name startswith /home/core/.ssh)

- macro: run_by_package_mgmt_binaries
  condition: proc.aname in (package_mgmt_binaries, needrestart)

- list: ssl_mgmt_binaries
  items: [ca-certificates]

- list: dhcp_binaries
  items: [dhclient, dhclient-script, 11-dhclient]

# A canonical set of processes that run other programs with different
# privileges or as a different user.
- list: userexec_binaries
  items: [sudo, su, suexec, critical-stack, dzdo]

- list: known_setuid_binaries
  items: [
    sshd, dbus-daemon-lau, ping, ping6, critical-stack-, pmmcli,
    filemng, PassengerAgent, bwrap, osdetect, nginxmng, sw-engine-fpm,
    start-stop-daem
    ]

- list: user_mgmt_binaries
  items: [login_binaries, passwd_binaries, shadowutils_binaries]

- list: dev_creation_binaries
  items: [blkid, rename_device, update_engine, sgdisk]

- list: hids_binaries
  items: [aide, aide.wrapper, update-aide.con, logcheck, syslog-summary, osqueryd, ossec-syscheckd]

- list: vpn_binaries
  items: [openvpn]

- list: nomachine_binaries
  items: [nxexec, nxnode.bin, nxserver.bin, nxclient.bin]

- macro: system_procs
  condition: proc.name in (coreutils_binaries, user_mgmt_binaries)

- list: mail_binaries
  items: [
    sendmail, sendmail-msp, postfix, procmail, exim4,
    pickup, showq, mailq, dovecot, imap-login, imap,
    mailmng-core, pop3-login, dovecot-lda, pop3
    ]

- list: mail_config_binaries
  items: [
    update_conf, parse_mc, makemap_hash, newaliases, update_mk, update_tlsm4,
    update_db, update_mc, ssmtp.postinst, mailq, postalias, postfix.config.,
    postfix.config, postfix-script, postconf
    ]

- list: sensitive_file_names
  items: [/etc/shadow, /etc/sudoers, /etc/pam.conf, /etc/security/pwquality.conf]

- list: sensitive_directory_names
  items: [/, /etc, /etc/, /root, /root/]

- macro: sensitive_files
  condition: >
    fd.name startswith /etc and
    (fd.name in (sensitive_file_names)
     or fd.directory in (/etc/sudoers.d, /etc/pam.d))

# Indicates that the process is new. Currently detected using time
# since process was started, using a threshold of 5 seconds.
- macro: proc_is_new
  condition: proc.duration <= 5000000000

# Network
- macro: inbound
  condition: >
    (((evt.type in (accept,listen) and evt.dir=<) or
      (evt.type in (recvfrom,recvmsg) and evt.dir=< and
       fd.l4proto != tcp and fd.connected=false and fd.name_changed=true)) and
     (fd.typechar = 4 or fd.typechar = 6) and
     (fd.ip != "0.0.0.0" and fd.net != "127.0.0.0/8") and
     (evt.rawres >= 0 or evt.res = EINPROGRESS))

# RFC1918 addresses were assigned for private network usage
- list: rfc_1918_addresses
  items: ['"10.0.0.0/8"', '"172.16.0.0/12"', '"192.168.0.0/16"']

- macro: outbound
  condition: >
    (((evt.type = connect and evt.dir=<) or
      (evt.type in (sendto,sendmsg) and evt.dir=< and
       fd.l4proto != tcp and fd.connected=false and fd.name_changed=true)) and
     (fd.typechar = 4 or fd.typechar = 6) and
     (fd.ip != "0.0.0.0" and fd.net != "127.0.0.0/8" and not fd.snet in (rfc_1918_addresses)) and
     (evt.rawres >= 0 or evt.res = EINPROGRESS))

# Very similar to inbound/outbound, but combines the tests together
# for efficiency.
- macro: inbound_outbound
  condition: >
    (((evt.type in (accept,listen,connect) and evt.dir=<)) or
     (fd.typechar = 4 or fd.typechar = 6) and
     (fd.ip != "0.0.0.0" and fd.net != "127.0.0.0/8") and
     (evt.rawres >= 0 or evt.res = EINPROGRESS))

- macro: ssh_port
  condition: fd.sport=22

# In a local/user rules file, you could override this macro to
# enumerate the servers for which ssh connections are allowed. For
# example, you might have a ssh gateway host for which ssh connections
# are allowed.
#
# In the main falco rules file, there isn't any way to know the
# specific hosts for which ssh access is allowed, so this macro just
# repeats ssh_port, which effectively allows ssh from all hosts. In
# the overridden macro, the condition would look something like
# "fd.sip="a.b.c.d" or fd.sip="e.f.g.h" or ..."
- macro: allowed_ssh_hosts
  condition: ssh_port

- rule: Disallowed SSH Connection
  desc: Detect any new ssh connection to a host other than those in an allowed group of hosts
  condition: (inbound_outbound) and ssh_port and not allowed_ssh_hosts
  output: Disallowed SSH Connection (command=%proc.cmdline connection=%fd.name user=%user.name container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [network, mitre_remote_service]

# These rules and supporting macros are more of an example for how to
# use the fd.*ip and fd.*ip.name fields to match connection
# information against ips, netmasks, and complete domain names.
#
# To use this rule, you should modify consider_all_outbound_conns and
# populate allowed_{source,destination}_{ipaddrs,networks,domains} with the
# values that make sense for your environment.
- macro: consider_all_outbound_conns
  condition: (never_true)

# Note that this can be either individual IPs or netmasks
- list: allowed_outbound_destination_ipaddrs
  items: ['"127.0.0.1"', '"8.8.8.8"']

- list: allowed_outbound_destination_networks
  items: ['"127.0.0.1/8"']

- list: allowed_outbound_destination_domains
  items: [google.com, www.yahoo.com]

- rule: Unexpected outbound connection destination
  desc: Detect any outbound connection to a destination outside of an allowed set of ips, networks, or domain names
  condition: >
    consider_all_outbound_conns and outbound and not
    ((fd.sip in (allowed_outbound_destination_ipaddrs)) or
     (fd.snet in (allowed_outbound_destination_networks)) or
     (fd.sip.name in (allowed_outbound_destination_domains)))
  output: Disallowed outbound connection destination (command=%proc.cmdline connection=%fd.name user=%user.name container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [network]

- macro: consider_all_inbound_conns
  condition: (never_true)

- list: allowed_inbound_source_ipaddrs
  items: ['"127.0.0.1"']

- list: allowed_inbound_source_networks
  items: ['"127.0.0.1/8"', '"10.0.0.0/8"']

- list: allowed_inbound_source_domains
  items: [google.com]

- rule: Unexpected inbound connection source
  desc: Detect any inbound connection from a source outside of an allowed set of ips, networks, or domain names
  condition: >
    consider_all_inbound_conns and inbound and not
    ((fd.cip in (allowed_inbound_source_ipaddrs)) or
     (fd.cnet in (allowed_inbound_source_networks)) or
     (fd.cip.name in (allowed_inbound_source_domains)))
  output: Disallowed inbound connection source (command=%proc.cmdline connection=%fd.name user=%user.name container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [network]

- list: bash_config_filenames
  items: [.bashrc, .bash_profile, .bash_history, .bash_login, .bash_logout, .inputrc, .profile]

- list: bash_config_files
  items: [/etc/profile, /etc/bashrc]

# Covers both csh and tcsh
- list: csh_config_filenames
  items: [.cshrc, .login, .logout, .history, .tcshrc, .cshdirs]

- list: csh_config_files
  items: [/etc/csh.cshrc, /etc/csh.login]

- list: zsh_config_filenames
  items: [.zshenv, .zprofile, .zshrc, .zlogin, .zlogout]

- list: shell_config_filenames
  items: [bash_config_filenames, csh_config_filenames, zsh_config_filenames]

- list: shell_config_files
  items: [bash_config_files, csh_config_files]

- list: shell_config_directories
  items: [/etc/zsh]

- rule: Modify Shell Configuration File
  desc: Detect attempt to modify shell configuration files
  condition: >
    open_write and
    (fd.filename in (shell_config_filenames) or
     fd.name in (shell_config_files) or
     fd.directory in (shell_config_directories)) and
    not proc.name in (shell_binaries)
  output: >
    a shell configuration file has been modified (user=%user.name command=%proc.cmdline file=%fd.name container_id=%container.id image=%container.image.repository)
  priority:
    WARNING
  tag: [file, mitre_persistence]

# This rule is not enabled by default, as there are many legitimate
# readers of shell config files. If you want to enable it, modify the
# following macro.

- macro: consider_shell_config_reads
  condition: (never_true)

- rule: Read Shell Configuration File
  desc: Detect attempts to read shell configuration files by non-shell programs
  condition: >
    open_read and
    consider_shell_config_reads and
    (fd.filename in (shell_config_filenames) or
     fd.name in (shell_config_files) or
     fd.directory in (shell_config_directories)) and
    (not proc.name in (shell_binaries))
  output: >
    a shell configuration file was read by a non-shell program (user=%user.name command=%proc.cmdline file=%fd.name container_id=%container.id image=%container.image.repository)
  priority:
    WARNING
  tag: [file, mitre_discovery]

- macro: consider_all_cron_jobs
  condition: (never_true)

- rule: Schedule Cron Jobs
  desc: Detect cron jobs scheduled
  condition: >
    consider_all_cron_jobs and
    ((open_write and fd.name startswith /etc/cron) or
     (spawned_process and proc.name = "crontab"))
  output: >
    Cron jobs were scheduled to run (user=%user.name command=%proc.cmdline
    file=%fd.name container_id=%container.id container_name=%container.name image=%container.image.repository:%container.image.tag)
  priority:
    NOTICE
  tag: [file, mitre_persistence]

# Use this to test whether the event occurred within a container.

# When displaying container information in the output field, use
# %container.info, without any leading term (file=%fd.name
# %container.info user=%user.name, and not file=%fd.name
# container=%container.info user=%user.name). The output will change
# based on the context and whether or not -pk/-pm/-pc was specified on
# the command line.
- macro: container
  condition: (container.id != host)

- macro: container_started
  condition: >
    ((evt.type = container or
     (evt.type=execve and evt.dir=< and proc.vpid=1)) and
     container.image.repository != incomplete)

- macro: interactive
  condition: >
    ((proc.aname=sshd and proc.name != sshd) or
    proc.name=systemd-logind or proc.name=login)

- list: cron_binaries
  items: [anacron, cron, crond, crontab]

# https://github.com/liske/needrestart
- list: needrestart_binaries
  items: [needrestart, 10-dpkg, 20-rpm, 30-pacman]

# Possible scripts run by sshkit
- list: sshkit_script_binaries
  items: [10_etc_sudoers., 10_passwd_group]

- list: plesk_binaries
  items: [sw-engine, sw-engine-fpm, sw-engine-kv, filemng, f2bmng]

# System users that should never log into a system. Consider adding your own
# service users (e.g. 'apache' or 'mysqld') here.
- macro: system_users
  condition: user.name in (bin, daemon, games, lp, mail, nobody, sshd, sync, uucp, www-data)

# These macros will be removed soon. Only keeping them to maintain
# compatiblity with some widely used rules files.
# Begin Deprecated
- macro: parent_ansible_running_python
  condition: (proc.pname in (python, pypy, python3) and proc.pcmdline contains ansible)

- macro: parent_bro_running_python
  condition: (proc.pname=python and proc.cmdline contains /usr/share/broctl)

- macro: parent_python_running_denyhosts
  condition: >
    (proc.cmdline startswith "denyhosts.py /usr/bin/denyhosts.py" or
     (proc.pname=python and
     (proc.pcmdline contains /usr/sbin/denyhosts or
      proc.pcmdline contains /usr/local/bin/denyhosts.py)))

- macro: parent_python_running_sdchecks
  condition: >
    (proc.pname in (python, python2.7) and
    (proc.pcmdline contains /opt/draios/bin/sdchecks))

- macro: python_running_sdchecks
  condition: >
    (proc.name in (python, python2.7) and
    (proc.cmdline contains /opt/draios/bin/sdchecks))

- macro: parent_linux_image_upgrade_script
  condition: proc.pname startswith linux-image-

- macro: parent_java_running_echo
  condition: (proc.pname=java and proc.cmdline startswith "sh -c echo")

- macro: parent_scripting_running_builds
  condition: >
    (proc.pname in (php,php5-fpm,php-fpm7.1,python,ruby,ruby2.3,ruby2.1,node,conda) and (
       proc.cmdline startswith "sh -c git" or
       proc.cmdline startswith "sh -c date" or
       proc.cmdline startswith "sh -c /usr/bin/g++" or
       proc.cmdline startswith "sh -c /usr/bin/gcc" or
       proc.cmdline startswith "sh -c gcc" or
       proc.cmdline startswith "sh -c if type gcc" or
       proc.cmdline startswith "sh -c cd '/var/www/edi/';LC_ALL=en_US.UTF-8 git" or
       proc.cmdline startswith "sh -c /var/www/edi/bin/sftp.sh" or
       proc.cmdline startswith "sh -c /usr/src/app/crxlsx/bin/linux/crxlsx" or
       proc.cmdline startswith "sh -c make parent" or
       proc.cmdline startswith "node /jenkins/tools" or
       proc.cmdline startswith "sh -c '/usr/bin/node'" or
       proc.cmdline startswith "sh -c stty -a |" or
       proc.pcmdline startswith "node /opt/nodejs/bin/yarn" or
       proc.pcmdline startswith "node /usr/local/bin/yarn" or
       proc.pcmdline startswith "node /root/.config/yarn" or
       proc.pcmdline startswith "node /opt/yarn/bin/yarn.js"))


- macro: httpd_writing_ssl_conf
  condition: >
    (proc.pname=run-httpd and
     (proc.cmdline startswith "sed -ri" or proc.cmdline startswith "sed -i") and
     (fd.name startswith /etc/httpd/conf.d/ or fd.name startswith /etc/httpd/conf))

- macro: userhelper_writing_etc_security
  condition: (proc.name=userhelper and fd.name startswith /etc/security)

- macro: parent_Xvfb_running_xkbcomp
  condition: (proc.pname=Xvfb and proc.cmdline startswith 'sh -c "/usr/bin/xkbcomp"')

- macro: parent_nginx_running_serf
  condition: (proc.pname=nginx and proc.cmdline startswith "sh -c serf")

- macro: parent_node_running_npm
  condition: (proc.pcmdline startswith "node /usr/local/bin/npm" or
              proc.pcmdline startswith "node /usr/local/nodejs/bin/npm" or
              proc.pcmdline startswith "node /opt/rh/rh-nodejs6/root/usr/bin/npm")

- macro: parent_java_running_sbt
  condition: (proc.pname=java and proc.pcmdline contains sbt-launch.jar)

- list: known_container_shell_spawn_cmdlines
  items: []

- list: known_shell_spawn_binaries
  items: []

## End Deprecated

- macro: ansible_running_python
  condition: (proc.name in (python, pypy, python3) and proc.cmdline contains ansible)

- macro: python_running_chef
  condition: (proc.name=python and (proc.cmdline contains yum-dump.py or proc.cmdline="python /usr/bin/chef-monitor.py"))

- macro: python_running_denyhosts
  condition: >
    (proc.name=python and
    (proc.cmdline contains /usr/sbin/denyhosts or
     proc.cmdline contains /usr/local/bin/denyhosts.py))

# Qualys seems to run a variety of shell subprocesses, at various
# levels. This checks at a few levels without the cost of a full
# proc.aname, which traverses the full parent heirarchy.
- macro: run_by_qualys
  condition: >
    (proc.pname=qualys-cloud-ag or
     proc.aname[2]=qualys-cloud-ag or
     proc.aname[3]=qualys-cloud-ag or
     proc.aname[4]=qualys-cloud-ag)

- macro: run_by_sumologic_securefiles
  condition: >
    ((proc.cmdline="usermod -a -G sumologic_collector" or
      proc.cmdline="groupadd sumologic_collector") and
     (proc.pname=secureFiles.sh and proc.aname[2]=java))

- macro: run_by_yum
  condition: ((proc.pname=sh and proc.aname[2]=yum) or
              (proc.aname[2]=sh and proc.aname[3]=yum))

- macro: run_by_ms_oms
  condition: >
    (proc.aname[3] startswith omsagent- or
     proc.aname[3] startswith scx-)

- macro: run_by_google_accounts_daemon
  condition: >
    (proc.aname[1] startswith google_accounts or
     proc.aname[2] startswith google_accounts or
     proc.aname[3] startswith google_accounts)

# Chef is similar.
- macro: run_by_chef
  condition: (proc.aname[2]=chef_command_wr or proc.aname[3]=chef_command_wr or
              proc.aname[2]=chef-client or proc.aname[3]=chef-client or
              proc.name=chef-client)

- macro: run_by_adclient
  condition: (proc.aname[2]=adclient or proc.aname[3]=adclient or proc.aname[4]=adclient)

- macro: run_by_centrify
  condition: (proc.aname[2]=centrify or proc.aname[3]=centrify or proc.aname[4]=centrify)

- macro: run_by_puppet
  condition: (proc.aname[2]=puppet or proc.aname[3]=puppet)

# Also handles running semi-indirectly via scl
- macro: run_by_foreman
  condition: >
    (user.name=foreman and
     (proc.pname in (rake, ruby, scl) and proc.aname[5] in (tfm-rake,tfm-ruby)) or
     (proc.pname=scl and proc.aname[2] in (tfm-rake,tfm-ruby)))

- macro: java_running_sdjagent
  condition: proc.name=java and proc.cmdline contains sdjagent.jar

- macro: kubelet_running_loopback
  condition: (proc.pname=kubelet and proc.name=loopback)

- macro: python_mesos_marathon_scripting
  condition: (proc.pcmdline startswith "python3 /marathon-lb/marathon_lb.py")

- macro: splunk_running_forwarder
  condition: (proc.pname=splunkd and proc.cmdline startswith "sh -c /opt/splunkforwarder")

- macro: parent_supervise_running_multilog
  condition: (proc.name=multilog and proc.pname=supervise)

- macro: supervise_writing_status
  condition: (proc.name in (supervise,svc) and fd.name startswith "/etc/sb/")

- macro: pki_realm_writing_realms
  condition: (proc.cmdline startswith "bash /usr/local/lib/pki/pki-realm" and fd.name startswith /etc/pki/realms)

- macro: htpasswd_writing_passwd
  condition: (proc.name=htpasswd and fd.name=/etc/nginx/.htpasswd)

- macro: lvprogs_writing_conf
  condition: >
    (proc.name in (dmeventd,lvcreate,pvscan) and
     (fd.name startswith /etc/lvm/archive or
      fd.name startswith /etc/lvm/backup or
      fd.name startswith /etc/lvm/cache))

- macro: ovsdb_writing_openvswitch
  condition: (proc.name=ovsdb-server and fd.directory=/etc/openvswitch)

- macro: perl_running_plesk
  condition: (proc.cmdline startswith "perl /opt/psa/admin/bin/plesk_agent_manager" or
              proc.pcmdline startswith "perl /opt/psa/admin/bin/plesk_agent_manager")

- macro: perl_running_updmap
  condition: (proc.cmdline startswith "perl /usr/bin/updmap")

- macro: perl_running_centrifydc
  condition: (proc.cmdline startswith "perl /usr/share/centrifydc")

- macro: runuser_reading_pam
  condition: (proc.name=runuser and fd.directory=/etc/pam.d)

- macro: parent_ucf_writing_conf
  condition: (proc.pname=ucf and proc.aname[2]=frontend)

- macro: consul_template_writing_conf
  condition: >
    ((proc.name=consul-template and fd.name startswith /etc/haproxy) or
     (proc.name=reload.sh and proc.aname[2]=consul-template and fd.name startswith /etc/ssl))

- macro: countly_writing_nginx_conf
  condition: (proc.cmdline startswith "nodejs /opt/countly/bin" and fd.name startswith /etc/nginx)

- list: ms_oms_binaries
  items: [omi.postinst, omsconfig.posti, scx.postinst, omsadmin.sh, omiagent]

- macro: ms_oms_writing_conf
  condition: >
    ((proc.name in (omiagent,omsagent,in_heartbeat_r*,omsadmin.sh,PerformInventor)
       or proc.pname in (ms_oms_binaries)
       or proc.aname[2] in (ms_oms_binaries))
     and (fd.name startswith /etc/opt/omi or fd.name startswith /etc/opt/microsoft/omsagent))

- macro: ms_scx_writing_conf
  condition: (proc.name in (GetLinuxOS.sh) and fd.name startswith /etc/opt/microsoft/scx)

- macro: azure_scripts_writing_conf
  condition: (proc.pname startswith "bash /var/lib/waagent/" and fd.name startswith /etc/azure)

- macro: azure_networkwatcher_writing_conf
  condition: (proc.name in (NetworkWatcherA) and fd.name=/etc/init.d/AzureNetworkWatcherAgent)

- macro: couchdb_writing_conf
  condition: (proc.name=beam.smp and proc.cmdline contains couchdb and fd.name startswith /etc/couchdb)

- macro: update_texmf_writing_conf
  condition: (proc.name=update-texmf and fd.name startswith /etc/texmf)

- macro: slapadd_writing_conf
  condition: (proc.name=slapadd and fd.name startswith /etc/ldap)

- macro: openldap_writing_conf
  condition: (proc.pname=run-openldap.sh and fd.name startswith /etc/openldap)

- macro: ucpagent_writing_conf
  condition: (proc.name=apiserver and container.image.repository=docker/ucp-agent and fd.name=/etc/authorization_config.cfg)

- macro: iscsi_writing_conf
  condition: (proc.name=iscsiadm and fd.name startswith /etc/iscsi)

- macro: istio_writing_conf
  condition: (proc.name=pilot-agent and fd.name startswith /etc/istio)

- macro: symantec_writing_conf
  condition: >
    ((proc.name=symcfgd and fd.name startswith /etc/symantec) or
     (proc.name=navdefutil and fd.name=/etc/symc-defutils.conf))

- macro: liveupdate_writing_conf
  condition: (proc.cmdline startswith "java LiveUpdate" and fd.name in (/etc/liveupdate.conf, /etc/Product.Catalog.JavaLiveUpdate))

- macro: rancher_agent
  condition: (proc.name=agent and container.image.repository contains "rancher/agent")

- macro: rancher_network_manager
  condition: (proc.name=rancher-bridge and container.image.repository contains "rancher/network-manager")

- macro: sosreport_writing_files
  condition: >
    (proc.name=urlgrabber-ext- and proc.aname[3]=sosreport and
     (fd.name startswith /etc/pkt/nssdb or fd.name startswith /etc/pki/nssdb))

- macro: pkgmgmt_progs_writing_pki
  condition: >
    (proc.name=urlgrabber-ext- and proc.pname in (yum, yum-cron, repoquery) and
     (fd.name startswith /etc/pkt/nssdb or fd.name startswith /etc/pki/nssdb))

- macro: update_ca_trust_writing_pki
  condition: (proc.pname=update-ca-trust and proc.name=trust and fd.name startswith /etc/pki)

- macro: brandbot_writing_os_release
  condition: proc.name=brandbot and fd.name=/etc/os-release

- macro: selinux_writing_conf
  condition: (proc.name in (semodule,genhomedircon,sefcontext_comp) and fd.name startswith /etc/selinux)

- list: veritas_binaries
  items: [vxconfigd, sfcache, vxclustadm, vxdctl, vxprint, vxdmpadm, vxdisk, vxdg, vxassist, vxtune]

- macro: veritas_driver_script
  condition: (proc.cmdline startswith "perl /opt/VRTSsfmh/bin/mh_driver.pl")

- macro: veritas_progs
  condition: (proc.name in (veritas_binaries) or veritas_driver_script)

- macro: veritas_writing_config
  condition: (veritas_progs and (fd.name startswith /etc/vx or fd.name startswith /etc/opt/VRTS or fd.name startswith /etc/vom))

- macro: nginx_writing_conf
  condition: (proc.name in (nginx,nginx-ingress-c,nginx-ingress) and (fd.name startswith /etc/nginx or fd.name startswith /etc/ingress-controller))

- macro: nginx_writing_certs
  condition: >
    (((proc.name=openssl and proc.pname=nginx-launch.sh) or proc.name=nginx-launch.sh) and fd.name startswith /etc/nginx/certs)

- macro: chef_client_writing_conf
  condition: (proc.pcmdline startswith "chef-client /opt/gitlab" and fd.name startswith /etc/gitlab)

- macro: centrify_writing_krb
  condition: (proc.name in (adjoin,addns) and fd.name startswith /etc/krb5)

- macro: cockpit_writing_conf
  condition: >
    ((proc.pname=cockpit-kube-la or proc.aname[2]=cockpit-kube-la)
     and fd.name startswith /etc/cockpit)

- macro: ipsec_writing_conf
  condition: (proc.name=start-ipsec.sh and fd.directory=/etc/ipsec)

- macro: exe_running_docker_save
  condition: (proc.cmdline startswith "exe /var/lib/docker" and proc.pname in (dockerd, docker))

# Ideally we'd have a length check here as well but sysdig
# filterchecks don't have operators like len()
- macro: sed_temporary_file
  condition: (proc.name=sed and fd.name startswith "/etc/sed")

- macro: python_running_get_pip
  condition: (proc.cmdline startswith "python get-pip.py")

- macro: python_running_ms_oms
  condition: (proc.cmdline startswith "python /var/lib/waagent/")

- macro: gugent_writing_guestagent_log
  condition: (proc.name=gugent and fd.name=GuestAgent.log)

- macro: dse_writing_tmp
  condition: (proc.name=dse-entrypoint and fd.name=/root/tmp__)

- macro: zap_writing_state
  condition: (proc.name=java and proc.cmdline contains "jar /zap" and fd.name startswith /root/.ZAP)

- macro: airflow_writing_state
  condition: (proc.name=airflow and fd.name startswith /root/airflow)

- macro: rpm_writing_root_rpmdb
  condition: (proc.name=rpm and fd.directory=/root/.rpmdb)

- macro: maven_writing_groovy
  condition: (proc.name=java and proc.cmdline contains "classpath /usr/local/apache-maven" and fd.name startswith /root/.groovy)

- macro: chef_writing_conf
  condition: (proc.name=chef-client and fd.name startswith /root/.chef)

- macro: kubectl_writing_state
  condition: (proc.name in (kubectl,oc) and fd.name startswith /root/.kube)

- macro: java_running_cassandra
  condition: (proc.name=java and proc.cmdline contains "cassandra.jar")

- macro: cassandra_writing_state
  condition: (java_running_cassandra and fd.directory=/root/.cassandra)

# Istio
- macro: galley_writing_state
  condition: (proc.name=galley and fd.name in (known_istio_files))

- list: known_istio_files
  items: [/healthready, /healthliveness]

- macro: calico_writing_state
  condition: (proc.name=kube-controller and fd.name startswith /status.json and k8s.pod.name startswith calico)

- list: repository_files
  items: [sources.list]

- list: repository_directories
  items: [/etc/apt/sources.list.d, /etc/yum.repos.d]

- macro: access_repositories
  condition: (fd.filename in (repository_files) or fd.directory in (repository_directories))

- macro: modify_repositories
  condition: (evt.arg.newpath pmatch (repository_directories))

- rule: Update Package Repository
  desc: Detect package repositories get updated
  condition: >
    ((open_write and access_repositories) or (modify and modify_repositories)) and not package_mgmt_procs
  output: >
    Repository files get updated (user=%user.name command=%proc.cmdline file=%fd.name newpath=%evt.arg.newpath container_id=%container.id image=%container.image.repository)
  priority:
    NOTICE
  tags: [filesystem, mitre_persistence]

- rule: Write below binary dir
  desc: an attempt to write to any file below a set of binary directories
  condition: >
    bin_dir and evt.dir = < and open_write
    and not package_mgmt_procs
    and not exe_running_docker_save
    and not python_running_get_pip
    and not python_running_ms_oms
  output: >
    File below a known binary directory opened for writing (user=%user.name
    command=%proc.cmdline file=%fd.name parent=%proc.pname pcmdline=%proc.pcmdline gparent=%proc.aname[2] container_id=%container.id image=%container.image.repository)
  priority: ERROR
  tags: [filesystem, mitre_persistence]

# If you'd like to generally monitor a wider set of directories on top
# of the ones covered by the rule Write below binary dir, you can use
# the following rule and lists.

- list: monitored_directories
  items: [/boot, /lib, /lib64, /usr/lib, /usr/local/lib, /usr/local/sbin, /usr/local/bin, /root/.ssh, /etc/cardserver]

# Until https://github.com/draios/sysdig/pull/1153, which fixes
# https://github.com/draios/sysdig/issues/1152, is widely available,
# we can't use glob operators to match pathnames. Until then, we do a
# looser check to match ssh directories.
# When fixed, we will use "fd.name glob '/home/*/.ssh/*'"
- macro: user_ssh_directory
  condition: (fd.name startswith '/home' and fd.name contains '.ssh')

# google_accounts_(daemon)
- macro: google_accounts_daemon_writing_ssh
  condition: (proc.name=google_accounts and user_ssh_directory)

- macro: cloud_init_writing_ssh
  condition: (proc.name=cloud-init and user_ssh_directory)

- macro: mkinitramfs_writing_boot
  condition: (proc.pname in (mkinitramfs, update-initramf) and fd.directory=/boot)

- macro: monitored_dir
  condition: >
    (fd.directory in (monitored_directories)
     or user_ssh_directory)
    and not mkinitramfs_writing_boot

# Add conditions to this macro (probably in a separate file,
# overwriting this macro) to allow for specific combinations of
# programs writing below monitored directories.
#
# Its default value is an expression that always is false, which
# becomes true when the "not ..." in the rule is applied.
- macro: user_known_write_monitored_dir_conditions
  condition: (never_true)

- rule: Write below monitored dir
  desc: an attempt to write to any file below a set of binary directories
  condition: >
    evt.dir = < and open_write and monitored_dir
    and not package_mgmt_procs
    and not coreos_write_ssh_dir
    and not exe_running_docker_save
    and not python_running_get_pip
    and not python_running_ms_oms
    and not google_accounts_daemon_writing_ssh
    and not cloud_init_writing_ssh
    and not user_known_write_monitored_dir_conditions
  output: >
    File below a monitored directory opened for writing (user=%user.name
    command=%proc.cmdline file=%fd.name parent=%proc.pname pcmdline=%proc.pcmdline gparent=%proc.aname[2] container_id=%container.id image=%container.image.repository)
  priority: ERROR
  tags: [filesystem, mitre_persistence]

# This rule is disabled by default as many system management tools
# like ansible, etc can read these files/paths. Enable it using this macro.

- macro: consider_ssh_reads
  condition: (never_true)

- rule: Read ssh information
  desc: Any attempt to read files below ssh directories by non-ssh programs
  condition: >
    (consider_ssh_reads and
     (open_read or open_directory) and
     (user_ssh_directory or fd.name startswith /root/.ssh) and
     (not proc.name in (ssh_binaries)))
  output: >
    ssh-related file/directory read by non-ssh program (user=%user.name
    command=%proc.cmdline file=%fd.name parent=%proc.pname pcmdline=%proc.pcmdline container_id=%container.id image=%container.image.repository)
  priority: ERROR
  tags: [filesystem, mitre_discovery]

- list: safe_etc_dirs
  items: [/etc/cassandra, /etc/ssl/certs/java, /etc/logstash, /etc/nginx/conf.d, /etc/container_environment, /etc/hrmconfig, /etc/fluent/configs.d]

- macro: fluentd_writing_conf_files
  condition: (proc.name=start-fluentd and fd.name in (/etc/fluent/fluent.conf, /etc/td-agent/td-agent.conf))

- macro: qualys_writing_conf_files
  condition: (proc.name=qualys-cloud-ag and fd.name=/etc/qualys/cloud-agent/qagent-log.conf)

- macro: git_writing_nssdb
  condition: (proc.name=git-remote-http and fd.directory=/etc/pki/nssdb)

- macro: plesk_writing_keys
  condition: (proc.name in (plesk_binaries) and fd.name startswith /etc/sw/keys)

- macro: plesk_install_writing_apache_conf
  condition: (proc.cmdline startswith "bash -hB /usr/lib/plesk-9.0/services/webserver.apache configure"
              and fd.name="/etc/apache2/apache2.conf.tmp")

- macro: plesk_running_mktemp
  condition: (proc.name=mktemp and proc.aname[3] in (plesk_binaries))

- macro: networkmanager_writing_resolv_conf
  condition: proc.aname[2]=nm-dispatcher and fd.name=/etc/resolv.conf

- macro: add_shell_writing_shells_tmp
  condition: (proc.name=add-shell and fd.name=/etc/shells.tmp)

- macro: duply_writing_exclude_files
  condition: (proc.name=touch and proc.pcmdline startswith "bash /usr/bin/duply" and fd.name startswith "/etc/duply")

- macro: xmlcatalog_writing_files
  condition: (proc.name=update-xmlcatal and fd.directory=/etc/xml)

- macro: datadog_writing_conf
  condition: ((proc.cmdline startswith "python /opt/datadog-agent" or
               proc.cmdline startswith "entrypoint.sh /entrypoint.sh datadog start" or
               proc.cmdline startswith "agent.py /opt/datadog-agent")
              and fd.name startswith "/etc/dd-agent")

- macro: rancher_writing_conf
  condition: ((proc.name in (healthcheck, lb-controller, rancher-dns)) and
              (container.image.repository contains "rancher/healthcheck" or
               container.image.repository contains "rancher/lb-service-haproxy" or
               container.image.repository contains "rancher/dns") and
              (fd.name startswith "/etc/haproxy" or fd.name startswith "/etc/rancher-dns"))

- macro: rancher_writing_root
  condition: (proc.name=rancher-metadat and
              (container.image.repository contains "rancher/metadata" or container.image.repository contains "rancher/lb-service-haproxy") and
              fd.name startswith "/answers.json")

- macro: checkpoint_writing_state
  condition: (proc.name=checkpoint and
              container.image.repository contains "coreos/pod-checkpointer" and
              fd.name startswith "/etc/kubernetes")

- macro: jboss_in_container_writing_passwd
  condition: >
    ((proc.cmdline="run-java.sh /opt/jboss/container/java/run/run-java.sh"
      or proc.cmdline="run-java.sh /opt/run-java/run-java.sh")
     and container
     and fd.name=/etc/passwd)

- macro: curl_writing_pki_db
  condition: (proc.name=curl and fd.directory=/etc/pki/nssdb)

- macro: haproxy_writing_conf
  condition: ((proc.name in (update-haproxy-,haproxy_reload.) or proc.pname in (update-haproxy-,haproxy_reload,haproxy_reload.))
               and (fd.name=/etc/openvpn/client.map or fd.name startswith /etc/haproxy))

- macro: java_writing_conf
  condition: (proc.name=java and fd.name=/etc/.java/.systemPrefs/.system.lock)

- macro: rabbitmq_writing_conf
  condition: (proc.name=rabbitmq-server and fd.directory=/etc/rabbitmq)

- macro: rook_writing_conf
  condition: (proc.name=toolbox.sh and container.image.repository=rook/toolbox
              and fd.directory=/etc/ceph)

- macro: httpd_writing_conf_logs
  condition: (proc.name=httpd and fd.name startswith /etc/httpd/)

- macro: mysql_writing_conf
  condition: >
    ((proc.name in (start-mysql.sh, run-mysqld) or proc.pname=start-mysql.sh) and
     (fd.name startswith /etc/mysql or fd.directory=/etc/my.cnf.d))

- macro: redis_writing_conf
  condition: >
    (proc.name in (run-redis, redis-launcher.) and fd.name=/etc/redis.conf or fd.name startswith /etc/redis)

- macro: openvpn_writing_conf
  condition: (proc.name in (openvpn,openvpn-entrypo) and fd.name startswith /etc/openvpn)

- macro: php_handlers_writing_conf
  condition: (proc.name=php_handlers_co and fd.name=/etc/psa/php_versions.json)

- macro: sed_writing_temp_file
  condition: >
    ((proc.aname[3]=cron_start.sh and fd.name startswith /etc/security/sed) or
     (proc.name=sed and (fd.name startswith /etc/apt/sources.list.d/sed or
                         fd.name startswith /etc/apt/sed or
                         fd.name startswith /etc/apt/apt.conf.d/sed)))

- macro: cron_start_writing_pam_env
  condition: (proc.cmdline="bash /usr/sbin/start-cron" and fd.name=/etc/security/pam_env.conf)

# In some cases dpkg-reconfigur runs commands that modify /etc. Not
# putting the full set of package management programs yet.
- macro: dpkg_scripting
  condition: (proc.aname[2] in (dpkg-reconfigur, dpkg-preconfigu))

- macro: ufw_writing_conf
  condition: (proc.name=ufw and fd.directory=/etc/ufw)

- macro: calico_writing_conf
  condition: >
    (proc.name = calico-node and fd.name startswith /etc/calico)

- macro: prometheus_conf_writing_conf
  condition: (proc.name=prometheus-conf and fd.name startswith /etc/prometheus/config_out)

- macro: openshift_writing_conf
  condition: (proc.name=oc and fd.name startswith /etc/origin/node)

- macro: keepalived_writing_conf
  condition: (proc.name=keepalived and fd.name=/etc/keepalived/keepalived.conf)

- macro: etcd_manager_updating_dns
  condition: (container and proc.name=etcd-manager and fd.name=/etc/hosts)

# Add conditions to this macro (probably in a separate file,
# overwriting this macro) to allow for specific combinations of
# programs writing below specific directories below
# /etc. fluentd_writing_conf_files is a good example to follow, as it
# specifies both the program doing the writing as well as the specific
# files it is allowed to modify.
#
# In this file, it just takes one of the programs in the base macro
# and repeats it.

- macro: user_known_write_etc_conditions
  condition: proc.name=confd

# This is a placeholder for user to extend the whitelist for write below etc rule
- macro: user_known_write_below_etc_activities
  condition: (never_true)

- macro: write_etc_common
  condition: >
    etc_dir and evt.dir = < and open_write
    and proc_name_exists
    and not proc.name in (passwd_binaries, shadowutils_binaries, sysdigcloud_binaries,
                          package_mgmt_binaries, ssl_mgmt_binaries, dhcp_binaries,
                          dev_creation_binaries, shell_mgmt_binaries,
                          mail_config_binaries,
                          sshkit_script_binaries,
                          ldconfig.real, ldconfig, confd, gpg, insserv,
                          apparmor_parser, update-mime, tzdata.config, tzdata.postinst,
                          systemd, systemd-machine, systemd-sysuser,
                          debconf-show, rollerd, bind9.postinst, sv,
                          gen_resolvconf., update-ca-certi, certbot, runsv,
                          qualys-cloud-ag, locales.postins, nomachine_binaries,
                          adclient, certutil, crlutil, pam-auth-update, parallels_insta,
                          openshift-launc, update-rc.d, puppet)
    and not proc.pname in (sysdigcloud_binaries, mail_config_binaries, hddtemp.postins, sshkit_script_binaries, locales.postins, deb_binaries, dhcp_binaries)
    and not fd.name pmatch (safe_etc_dirs)
    and not fd.name in (/etc/container_environment.sh, /etc/container_environment.json, /etc/motd, /etc/motd.svc)
    and not sed_temporary_file
    and not exe_running_docker_save
    and not ansible_running_python
    and not python_running_denyhosts
    and not fluentd_writing_conf_files
    and not user_known_write_etc_conditions
    and not run_by_centrify
    and not run_by_adclient
    and not qualys_writing_conf_files
    and not git_writing_nssdb
    and not plesk_writing_keys
    and not plesk_install_writing_apache_conf
    and not plesk_running_mktemp
    and not networkmanager_writing_resolv_conf
    and not run_by_chef
    and not add_shell_writing_shells_tmp
    and not duply_writing_exclude_files
    and not xmlcatalog_writing_files
    and not parent_supervise_running_multilog
    and not supervise_writing_status
    and not pki_realm_writing_realms
    and not htpasswd_writing_passwd
    and not lvprogs_writing_conf
    and not ovsdb_writing_openvswitch
    and not datadog_writing_conf
    and not curl_writing_pki_db
    and not haproxy_writing_conf
    and not java_writing_conf
    and not dpkg_scripting
    and not parent_ucf_writing_conf
    and not rabbitmq_writing_conf
    and not rook_writing_conf
    and not php_handlers_writing_conf
    and not sed_writing_temp_file
    and not cron_start_writing_pam_env
    and not httpd_writing_conf_logs
    and not mysql_writing_conf
    and not openvpn_writing_conf
    and not consul_template_writing_conf
    and not countly_writing_nginx_conf
    and not ms_oms_writing_conf
    and not ms_scx_writing_conf
    and not azure_scripts_writing_conf
    and not azure_networkwatcher_writing_conf
    and not couchdb_writing_conf
    and not update_texmf_writing_conf
    and not slapadd_writing_conf
    and not symantec_writing_conf
    and not liveupdate_writing_conf
    and not sosreport_writing_files
    and not selinux_writing_conf
    and not veritas_writing_config
    and not nginx_writing_conf
    and not nginx_writing_certs
    and not chef_client_writing_conf
    and not centrify_writing_krb
    and not cockpit_writing_conf
    and not ipsec_writing_conf
    and not httpd_writing_ssl_conf
    and not userhelper_writing_etc_security
    and not pkgmgmt_progs_writing_pki
    and not update_ca_trust_writing_pki
    and not brandbot_writing_os_release
    and not redis_writing_conf
    and not openldap_writing_conf
    and not ucpagent_writing_conf
    and not iscsi_writing_conf
    and not istio_writing_conf
    and not ufw_writing_conf
    and not calico_writing_conf
    and not prometheus_conf_writing_conf
    and not openshift_writing_conf
    and not keepalived_writing_conf
    and not rancher_writing_conf
    and not checkpoint_writing_state
    and not jboss_in_container_writing_passwd
    and not etcd_manager_updating_dns
    and not user_known_write_below_etc_activities

- rule: Write below etc
  desc: an attempt to write to any file below /etc
  condition: write_etc_common
  output: "File below /etc opened for writing (user=%user.name command=%proc.cmdline parent=%proc.pname pcmdline=%proc.pcmdline file=%fd.name program=%proc.name gparent=%proc.aname[2] ggparent=%proc.aname[3] gggparent=%proc.aname[4] container_id=%container.id image=%container.image.repository)"
  priority: ERROR
  tags: [filesystem, mitre_persistence]

- list: known_root_files
  items: [/root/.monit.state, /root/.auth_tokens, /root/.bash_history, /root/.ash_history, /root/.aws/credentials,
          /root/.viminfo.tmp, /root/.lesshst, /root/.bzr.log, /root/.gitconfig.lock, /root/.babel.json, /root/.localstack,
          /root/.node_repl_history, /root/.mongorc.js, /root/.dbshell, /root/.augeas/history, /root/.rnd, /root/.wget-hsts, /health, /exec.fifo]

- list: known_root_directories
  items: [/root/.oracle_jre_usage, /root/.ssh, /root/.subversion, /root/.nami]

- macro: known_root_conditions
  condition: (fd.name startswith /root/orcexec.
              or fd.name startswith /root/.m2
              or fd.name startswith /root/.npm
              or fd.name startswith /root/.pki
              or fd.name startswith /root/.ivy2
              or fd.name startswith /root/.config/Cypress
              or fd.name startswith /root/.config/pulse
              or fd.name startswith /root/.config/configstore
              or fd.name startswith /root/jenkins/workspace
              or fd.name startswith /root/.jenkins
              or fd.name startswith /root/.cache
              or fd.name startswith /root/.sbt
              or fd.name startswith /root/.java
              or fd.name startswith /root/.glide
              or fd.name startswith /root/.sonar
              or fd.name startswith /root/.v8flag
              or fd.name startswith /root/infaagent
              or fd.name startswith /root/.local/lib/python
              or fd.name startswith /root/.pm2
              or fd.name startswith /root/.gnupg
              or fd.name startswith /root/.pgpass
              or fd.name startswith /root/.theano
              or fd.name startswith /root/.gradle
              or fd.name startswith /root/.android
              or fd.name startswith /root/.ansible
              or fd.name startswith /root/.crashlytics
              or fd.name startswith /root/.dbus
              or fd.name startswith /root/.composer
              or fd.name startswith /root/.gconf
              or fd.name startswith /root/.nv
              or fd.name startswith /root/.local/share/jupyter
              or fd.name startswith /root/oradiag_root
              or fd.name startswith /root/workspace
              or fd.name startswith /root/jvm
              or fd.name startswith /root/.node-gyp)

# Add conditions to this macro (probably in a separate file,
# overwriting this macro) to allow for specific combinations of
# programs writing below specific directories below
# / or /root.
#
# In this file, it just takes one of the condition in the base macro
# and repeats it.
- macro: user_known_write_root_conditions
  condition: fd.name=/root/.bash_history

# This is a placeholder for user to extend the whitelist for write below root rule
- macro: user_known_write_below_root_activities
  condition: (never_true)

- rule: Write below root
  desc: an attempt to write to any file directly below / or /root
  condition: >
    root_dir and evt.dir = < and open_write
    and not fd.name in (known_root_files)
    and not fd.directory in (known_root_directories)
    and not exe_running_docker_save
    and not gugent_writing_guestagent_log
    and not dse_writing_tmp
    and not zap_writing_state
    and not airflow_writing_state
    and not rpm_writing_root_rpmdb
    and not maven_writing_groovy
    and not chef_writing_conf
    and not kubectl_writing_state
    and not cassandra_writing_state
    and not galley_writing_state
    and not calico_writing_state
    and not rancher_writing_root
    and not known_root_conditions
    and not user_known_write_root_conditions
    and not user_known_write_below_root_activities
  output: "File below / or /root opened for writing (user=%user.name command=%proc.cmdline parent=%proc.pname file=%fd.name program=%proc.name container_id=%container.id image=%container.image.repository)"
  priority: ERROR
  tags: [filesystem, mitre_persistence]

- macro: cmp_cp_by_passwd
  condition: proc.name in (cmp, cp) and proc.pname in (passwd, run-parts)

- rule: Read sensitive file trusted after startup
  desc: >
    an attempt to read any sensitive file (e.g. files containing user/password/authentication
    information) by a trusted program after startup. Trusted programs might read these files
    at startup to load initial state, but not afterwards.
  condition: sensitive_files and open_read and server_procs and not proc_is_new and proc.name!="sshd"
  output: >
    Sensitive file opened for reading by trusted program after startup (user=%user.name
    command=%proc.cmdline parent=%proc.pname file=%fd.name parent=%proc.pname gparent=%proc.aname[2] container_id=%container.id image=%container.image.repository)
  priority: WARNING
  tags: [filesystem, mitre_credential_access]

- list: read_sensitive_file_binaries
  items: [
    iptables, ps, lsb_release, check-new-relea, dumpe2fs, accounts-daemon, sshd,
    vsftpd, systemd, mysql_install_d, psql, screen, debconf-show, sa-update,
    pam-auth-update, pam-config, /usr/sbin/spamd, polkit-agent-he, lsattr, file, sosreport,
    scxcimservera, adclient, rtvscand, cockpit-session, userhelper, ossec-syscheckd
    ]

# Add conditions to this macro (probably in a separate file,
# overwriting this macro) to allow for specific combinations of
# programs accessing sensitive files.
# fluentd_writing_conf_files is a good example to follow, as it
# specifies both the program doing the writing as well as the specific
# files it is allowed to modify.
#
# In this file, it just takes one of the macros in the base rule
# and repeats it.

- macro: user_read_sensitive_file_conditions
  condition: cmp_cp_by_passwd

- rule: Read sensitive file untrusted
  desc: >
    an attempt to read any sensitive file (e.g. files containing user/password/authentication
    information). Exceptions are made for known trusted programs.
  condition: >
    sensitive_files and open_read
    and proc_name_exists
    and not proc.name in (user_mgmt_binaries, userexec_binaries, package_mgmt_binaries,
     cron_binaries, read_sensitive_file_binaries, shell_binaries, hids_binaries,
     vpn_binaries, mail_config_binaries, nomachine_binaries, sshkit_script_binaries,
     in.proftpd, mandb, salt-minion, postgres_mgmt_binaries)
    and not cmp_cp_by_passwd
    and not ansible_running_python
    and not proc.cmdline contains /usr/bin/mandb
    and not run_by_qualys
    and not run_by_chef
    and not run_by_google_accounts_daemon
    and not user_read_sensitive_file_conditions
    and not perl_running_plesk
    and not perl_running_updmap
    and not veritas_driver_script
    and not perl_running_centrifydc
    and not runuser_reading_pam
  output: >
    Sensitive file opened for reading by non-trusted program (user=%user.name program=%proc.name
    command=%proc.cmdline file=%fd.name parent=%proc.pname gparent=%proc.aname[2] ggparent=%proc.aname[3] gggparent=%proc.aname[4] container_id=%container.id image=%container.image.repository)
  priority: WARNING
  tags: [filesystem, mitre_credential_access, mitre_discovery]

- macro: amazon_linux_running_python_yum
  condition: >
    (proc.name = python and
     proc.pcmdline = "python -m amazon_linux_extras system_motd" and
     proc.cmdline startswith "python -c import yum;")

# Only let rpm-related programs write to the rpm database
- rule: Write below rpm database
  desc: an attempt to write to the rpm database by any non-rpm related program
  condition: >
    fd.name startswith /var/lib/rpm and open_write
    and not rpm_procs
    and not ansible_running_python
    and not python_running_chef
    and not exe_running_docker_save
    and not amazon_linux_running_python_yum
  output: "Rpm database opened for writing by a non-rpm program (command=%proc.cmdline file=%fd.name parent=%proc.pname pcmdline=%proc.pcmdline container_id=%container.id image=%container.image.repository)"
  priority: ERROR
  tags: [filesystem, software_mgmt, mitre_persistence]

- macro: postgres_running_wal_e
  condition: (proc.pname=postgres and proc.cmdline startswith "sh -c envdir /etc/wal-e.d/env /usr/local/bin/wal-e")

- macro: redis_running_prepost_scripts
  condition: (proc.aname[2]=redis-server and (proc.cmdline contains "redis-server.post-up.d" or proc.cmdline contains "redis-server.pre-up.d"))

- macro: rabbitmq_running_scripts
  condition: >
    (proc.pname=beam.smp and
    (proc.cmdline startswith "sh -c exec ps" or
     proc.cmdline startswith "sh -c exec inet_gethost" or
     proc.cmdline= "sh -s unix:cmd" or
     proc.cmdline= "sh -c exec /bin/sh -s unix:cmd 2>&1"))

- macro: rabbitmqctl_running_scripts
  condition: (proc.aname[2]=rabbitmqctl and proc.cmdline startswith "sh -c ")

- macro: run_by_appdynamics
  condition: (proc.pname=java and proc.pcmdline startswith "java -jar -Dappdynamics")

- rule: DB program spawned process
  desc: >
    a database-server related program spawned a new process other than itself.
    This shouldn\'t occur and is a follow on from some SQL injection attacks.
  condition: >
    proc.pname in (db_server_binaries)
    and spawned_process
    and not proc.name in (db_server_binaries)
    and not postgres_running_wal_e
  output: >
    Database-related program spawned process other than itself (user=%user.name
    program=%proc.cmdline parent=%proc.pname container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [process, database, mitre_execution]

- rule: Modify binary dirs
  desc: an attempt to modify any file below a set of binary directories.
  condition: (bin_dir_rename) and modify and not package_mgmt_procs and not exe_running_docker_save
  output: >
    File below known binary directory renamed/removed (user=%user.name command=%proc.cmdline
    pcmdline=%proc.pcmdline operation=%evt.type file=%fd.name %evt.args container_id=%container.id image=%container.image.repository)
  priority: ERROR
  tags: [filesystem, mitre_persistence]

- rule: Mkdir binary dirs
  desc: an attempt to create a directory below a set of binary directories.
  condition: mkdir and bin_dir_mkdir and not package_mgmt_procs
  output: >
    Directory below known binary directory created (user=%user.name
    command=%proc.cmdline directory=%evt.arg.path container_id=%container.id image=%container.image.repository)
  priority: ERROR
  tags: [filesystem, mitre_persistence]

# This list allows for easy additions to the set of commands allowed
# to change thread namespace without having to copy and override the
# entire change thread namespace rule.
- list: user_known_change_thread_namespace_binaries
  items: []

- macro: user_known_change_thread_namespace_activities
  condition: (never_true)

- list: network_plugin_binaries
  items: [aws-cni, azure-vnet]

- macro: calico_node
  condition: (container.image.repository endswith calico/node and proc.name=calico-node)

- macro: weaveworks_scope
  condition: (container.image.repository endswith weaveworks/scope and proc.name=scope)

- rule: Change thread namespace
  desc: >
    an attempt to change a program/thread\'s namespace (commonly done
    as a part of creating a container) by calling setns.
  condition: >
    evt.type = setns
    and not proc.name in (docker_binaries, k8s_binaries, lxd_binaries, sysdigcloud_binaries,
                          sysdig, nsenter, calico, oci-umount, network_plugin_binaries)
    and not proc.name in (user_known_change_thread_namespace_binaries)
    and not proc.name startswith "runc"
    and not proc.cmdline startswith "containerd"
    and not proc.pname in (sysdigcloud_binaries)
    and not python_running_sdchecks
    and not java_running_sdjagent
    and not kubelet_running_loopback
    and not rancher_agent
    and not rancher_network_manager
    and not calico_node
    and not weaveworks_scope
    and not user_known_change_thread_namespace_activities
  output: >
    Namespace change (setns) by unexpected program (user=%user.name command=%proc.cmdline
    parent=%proc.pname %container.info container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [process]

# The binaries in this list and their descendents are *not* allowed
# spawn shells. This includes the binaries spawning shells directly as
# well as indirectly. For example, apache -> php/perl for
# mod_{php,perl} -> some shell is also not allowed, because the shell
# has apache as an ancestor.

- list: protected_shell_spawning_binaries
  items: [
    http_server_binaries, db_server_binaries, nosql_server_binaries, mail_binaries,
    fluentd, flanneld, splunkd, consul, smbd, runsv, PM2
    ]

- macro: parent_java_running_zookeeper
  condition: (proc.pname=java and proc.pcmdline contains org.apache.zookeeper.server)

- macro: parent_java_running_kafka
  condition: (proc.pname=java and proc.pcmdline contains kafka.Kafka)

- macro: parent_java_running_elasticsearch
  condition: (proc.pname=java and proc.pcmdline contains org.elasticsearch.bootstrap.Elasticsearch)

- macro: parent_java_running_activemq
  condition: (proc.pname=java and proc.pcmdline contains activemq.jar)

- macro: parent_java_running_cassandra
  condition: (proc.pname=java and (proc.pcmdline contains "-Dcassandra.config.loader" or proc.pcmdline contains org.apache.cassandra.service.CassandraDaemon))

- macro: parent_java_running_jboss_wildfly
  condition: (proc.pname=java and proc.pcmdline contains org.jboss)

- macro: parent_java_running_glassfish
  condition: (proc.pname=java and proc.pcmdline contains com.sun.enterprise.glassfish)

- macro: parent_java_running_hadoop
  condition: (proc.pname=java and proc.pcmdline contains org.apache.hadoop)

- macro: parent_java_running_datastax
  condition: (proc.pname=java and proc.pcmdline contains com.datastax)

- macro: nginx_starting_nginx
  condition: (proc.pname=nginx and proc.cmdline contains "/usr/sbin/nginx -c /etc/nginx/nginx.conf")

- macro: nginx_running_aws_s3_cp
  condition: (proc.pname=nginx and proc.cmdline startswith "sh -c /usr/local/bin/aws s3 cp")

- macro: consul_running_net_scripts
  condition: (proc.pname=consul and (proc.cmdline startswith "sh -c curl" or proc.cmdline startswith "sh -c nc"))

- macro: consul_running_alert_checks
  condition: (proc.pname=consul and proc.cmdline startswith "sh -c /bin/consul-alerts")

- macro: serf_script
  condition: (proc.cmdline startswith "sh -c serf")

- macro: check_process_status
  condition: (proc.cmdline startswith "sh -c kill -0 ")

# In some cases, you may want to consider node processes run directly
# in containers as protected shell spawners. Examples include using
# pm2-docker or pm2 start some-app.js --no-daemon-mode as the direct
# entrypoint of the container, and when the node app is a long-lived
# server using something like express.
#
# However, there are other uses of node related to build pipelines for
# which node is not really a server but instead a general scripting
# tool. In these cases, shells are very likely and in these cases you
# don't want to consider node processes protected shell spawners.
#
# We have to choose one of these cases, so we consider node processes
# as unprotected by default. If you want to consider any node process
# run in a container as a protected shell spawner, override the below
# macro to remove the "never_true" clause, which allows it to take effect.
- macro: possibly_node_in_container
  condition: (never_true and (proc.pname=node and proc.aname[3]=docker-containe))

# Similarly, you may want to consider any shell spawned by apache
# tomcat as suspect. The famous apache struts attack (CVE-2017-5638)
# could be exploited to do things like spawn shells.
#
# However, many applications *do* use tomcat to run arbitrary shells,
# as a part of build pipelines, etc.
#
# Like for node, we make this case opt-in.
- macro: possibly_parent_java_running_tomcat
  condition: (never_true and proc.pname=java and proc.pcmdline contains org.apache.catalina.startup.Bootstrap)

- macro: protected_shell_spawner
  condition: >
    (proc.aname in (protected_shell_spawning_binaries)
    or parent_java_running_zookeeper
    or parent_java_running_kafka
    or parent_java_running_elasticsearch
    or parent_java_running_activemq
    or parent_java_running_cassandra
    or parent_java_running_jboss_wildfly
    or parent_java_running_glassfish
    or parent_java_running_hadoop
    or parent_java_running_datastax
    or possibly_parent_java_running_tomcat
    or possibly_node_in_container)

- list: mesos_shell_binaries
  items: [mesos-docker-ex, mesos-slave, mesos-health-ch]

# Note that runsv is both in protected_shell_spawner and the
# exclusions by pname. This means that runsv can itself spawn shells
# (the ./run and ./finish scripts), but the processes runsv can not
# spawn shells.
- rule: Run shell untrusted
  desc: an attempt to spawn a shell below a non-shell application. Specific applications are monitored.
  condition: >
    spawned_process
    and shell_procs
    and proc.pname exists
    and protected_shell_spawner
    and not proc.pname in (shell_binaries, gitlab_binaries, cron_binaries, user_known_shell_spawn_binaries,
                           needrestart_binaries,
                           mesos_shell_binaries,
                           erl_child_setup, exechealthz,
                           PM2, PassengerWatchd, c_rehash, svlogd, logrotate, hhvm, serf,
                           lb-controller, nvidia-installe, runsv, statsite, erlexec)
    and not proc.cmdline in (known_shell_spawn_cmdlines)
    and not proc.aname in (unicorn_launche)
    and not consul_running_net_scripts
    and not consul_running_alert_checks
    and not nginx_starting_nginx
    and not nginx_running_aws_s3_cp
    and not run_by_package_mgmt_binaries
    and not serf_script
    and not check_process_status
    and not run_by_foreman
    and not python_mesos_marathon_scripting
    and not splunk_running_forwarder
    and not postgres_running_wal_e
    and not redis_running_prepost_scripts
    and not rabbitmq_running_scripts
    and not rabbitmqctl_running_scripts
    and not run_by_appdynamics
    and not user_shell_container_exclusions
  output: >
    Shell spawned by untrusted binary (user=%user.name shell=%proc.name parent=%proc.pname
    cmdline=%proc.cmdline pcmdline=%proc.pcmdline gparent=%proc.aname[2] ggparent=%proc.aname[3]
    aname[4]=%proc.aname[4] aname[5]=%proc.aname[5] aname[6]=%proc.aname[6] aname[7]=%proc.aname[7] container_id=%container.id image=%container.image.repository)
  priority: DEBUG
  tags: [shell, mitre_execution]

- macro: allowed_openshift_registry_root
  condition: >
    (container.image.repository startswith openshift3/ or
     container.image.repository startswith registry.redhat.io/openshift3/ or
     container.image.repository startswith registry.access.redhat.com/openshift3/)

# Source: https://docs.openshift.com/enterprise/3.2/install_config/install/disconnected_install.html
- macro: openshift_image
  condition: >
    (allowed_openshift_registry_root and
      (container.image.repository endswith /logging-deployment or
       container.image.repository endswith /logging-elasticsearch or
       container.image.repository endswith /logging-kibana or
       container.image.repository endswith /logging-fluentd or
       container.image.repository endswith /logging-auth-proxy or
       container.image.repository endswith /metrics-deployer or
       container.image.repository endswith /metrics-hawkular-metrics or
       container.image.repository endswith /metrics-cassandra or
       container.image.repository endswith /metrics-heapster or
       container.image.repository endswith /ose-haproxy-router or
       container.image.repository endswith /ose-deployer or
       container.image.repository endswith /ose-sti-builder or
       container.image.repository endswith /ose-docker-builder or
       container.image.repository endswith /ose-pod or
       container.image.repository endswith /ose-node or
       container.image.repository endswith /ose-docker-registry or
       container.image.repository endswith /prometheus-node-exporter or
       container.image.repository endswith /image-inspector))

# These images are allowed both to run with --privileged and to mount
# sensitive paths from the host filesystem.
#
# NOTE: This list is only provided for backwards compatibility with
# older local falco rules files that may have been appending to
# trusted_images. To make customizations, it's better to add images to
# either privileged_images or falco_sensitive_mount_images.
- list: trusted_images
  items: []

# NOTE: This macro is only provided for backwards compatibility with
# older local falco rules files that may have been appending to
# trusted_images. To make customizations, it's better to add containers to
# user_trusted_containers, user_privileged_containers or user_sensitive_mount_containers.
- macro: trusted_containers
  condition: (container.image.repository in (trusted_images))

# Add conditions to this macro (probably in a separate file,
# overwriting this macro) to specify additional containers that are
# trusted and therefore allowed to run privileged *and* with sensitive
# mounts.
#
# Like trusted_images, this is deprecated in favor of
# user_privileged_containers and user_sensitive_mount_containers and
# is only provided for backwards compatibility.
#
# In this file, it just takes one of the images in trusted_containers
# and repeats it.
- macro: user_trusted_containers
  condition: (container.image.repository endswith sysdig/agent)

- list: sematext_images
  items: [docker.io/sematext/sematext-agent-docker, docker.io/sematext/agent, docker.io/sematext/logagent,
          registry.access.redhat.com/sematext/sematext-agent-docker,
          registry.access.redhat.com/sematext/agent,
          registry.access.redhat.com/sematext/logagent]

# These container images are allowed to run with --privileged
- list: falco_privileged_images
  items: [
    docker.io/sysdig/agent, docker.io/sysdig/falco, docker.io/sysdig/sysdig,
    gcr.io/google_containers/kube-proxy, docker.io/calico/node,
    docker.io/rook/toolbox, docker.io/cloudnativelabs/kube-router, docker.io/mesosphere/mesos-slave,
    docker.io/docker/ucp-agent, sematext_images, k8s.gcr.io/kube-proxy
    ]

- macro: falco_privileged_containers
  condition: (openshift_image or
              user_trusted_containers or
              container.image.repository in (trusted_images) or
              container.image.repository in (falco_privileged_images) or
              container.image.repository startswith istio/proxy_ or
              container.image.repository startswith quay.io/sysdig)

# Add conditions to this macro (probably in a separate file,
# overwriting this macro) to specify additional containers that are
# allowed to run privileged
#
# In this file, it just takes one of the images in falco_privileged_images
# and repeats it.
- macro: user_privileged_containers
  condition: (container.image.repository endswith sysdig/agent)

- list: rancher_images
  items: [
    rancher/network-manager, rancher/dns, rancher/agent,
    rancher/lb-service-haproxy, rancher/metadata, rancher/healthcheck
  ]

# These container images are allowed to mount sensitive paths from the
# host filesystem.
- list: falco_sensitive_mount_images
  items: [
    docker.io/sysdig/agent, docker.io/sysdig/falco, docker.io/sysdig/sysdig,
    gcr.io/google_containers/hyperkube,
    gcr.io/google_containers/kube-proxy, docker.io/calico/node,
    docker.io/rook/toolbox, docker.io/cloudnativelabs/kube-router, docker.io/consul,
    docker.io/datadog/docker-dd-agent, docker.io/datadog/agent, docker.io/docker/ucp-agent, docker.io/gliderlabs/logspout,
    docker.io/netdata/netdata, docker.io/google/cadvisor, docker.io/prom/node-exporter
    ]

- macro: falco_sensitive_mount_containers
  condition: (user_trusted_containers or
              container.image.repository in (trusted_images) or
              container.image.repository in (falco_sensitive_mount_images) or
              container.image.repository startswith quay.io/sysdig)

# These container images are allowed to run with hostnetwork=true
- list: falco_hostnetwork_images
  items: []

# Add conditions to this macro (probably in a separate file,
# overwriting this macro) to specify additional containers that are
# allowed to perform sensitive mounts.
#
# In this file, it just takes one of the images in falco_sensitive_mount_images
# and repeats it.
- macro: user_sensitive_mount_containers
  condition: (container.image.repository = docker.io/sysdig/agent)

- rule: Launch Privileged Container
  desc: Detect the initial process started in a privileged container. Exceptions are made for known trusted images.
  condition: >
    container_started and container
    and container.privileged=true
    and not falco_privileged_containers
    and not user_privileged_containers
  output: Privileged container started (user=%user.name command=%proc.cmdline %container.info image=%container.image.repository:%container.image.tag)
  priority: INFO
  tags: [container, cis, mitre_privilege_escalation, mitre_lateral_movement]

# For now, only considering a full mount of /etc as
# sensitive. Ideally, this would also consider all subdirectories
# below /etc as well, but the globbing mechanism used by sysdig
# doesn't allow exclusions of a full pattern, only single characters.
- macro: sensitive_mount
  condition: (container.mount.dest[/proc*] != "N/A" or
              container.mount.dest[/var/run/docker.sock] != "N/A" or
              container.mount.dest[/var/lib/kubelet] != "N/A" or
              container.mount.dest[/var/lib/kubelet/pki] != "N/A" or
              container.mount.dest[/] != "N/A" or
              container.mount.dest[/etc] != "N/A" or
              container.mount.dest[/root*] != "N/A")

# The steps libcontainer performs to set up the root program for a container are:
# - clone + exec self to a program runc:[0:PARENT]
# - clone a program runc:[1:CHILD] which sets up all the namespaces
# - clone a second program runc:[2:INIT] + exec to the root program.
#   The parent of runc:[2:INIT] is runc:0:PARENT]
# As soon as 1:CHILD is created, 0:PARENT exits, so there's a race
#   where at the time 2:INIT execs the root program, 0:PARENT might have
#   already exited, or might still be around. So we handle both.
# We also let runc:[1:CHILD] count as the parent process, which can occur
# when we lose events and lose track of state.

- macro: container_entrypoint
  condition: (not proc.pname exists or proc.pname in (runc:[0:PARENT], runc:[1:CHILD], runc, docker-runc, exe))

- rule: Launch Sensitive Mount Container
  desc: >
    Detect the initial process started by a container that has a mount from a sensitive host directory
    (i.e. /proc). Exceptions are made for known trusted images.
  condition: >
    container_started and container
    and sensitive_mount
    and not falco_sensitive_mount_containers
    and not user_sensitive_mount_containers
  output: Container with sensitive mount started (user=%user.name command=%proc.cmdline %container.info image=%container.image.repository:%container.image.tag mounts=%container.mounts)
  priority: INFO
  tags: [container, cis, mitre_lateral_movement]

# In a local/user rules file, you could override this macro to
# explicitly enumerate the container images that you want to run in
# your environment. In this main falco rules file, there isn't any way
# to know all the containers that can run, so any container is
# allowed, by using a filter that is guaranteed to evaluate to true.
# In the overridden macro, the condition would look something like
# (container.image.repository = vendor/container-1 or
# container.image.repository = vendor/container-2 or ...)

- macro: allowed_containers
  condition: (container.id exists)

- rule: Launch Disallowed Container
  desc: >
    Detect the initial process started by a container that is not in a list of allowed containers.
  condition: container_started and container and not allowed_containers
  output: Container started and not in allowed list (user=%user.name command=%proc.cmdline %container.info image=%container.image.repository:%container.image.tag)
  priority: WARNING
  tags: [container, mitre_lateral_movement]

# Anything run interactively by root
# - condition: evt.type != switch and user.name = root and proc.name != sshd and interactive
#  output: "Interactive root (%user.name %proc.name %evt.dir %evt.type %evt.args %fd.name)"
#  priority: WARNING

- rule: System user interactive
  desc: an attempt to run interactive commands by a system (i.e. non-login) user
  condition: spawned_process and system_users and interactive
  output: "System user ran an interactive command (user=%user.name command=%proc.cmdline container_id=%container.id image=%container.image.repository)"
  priority: INFO
  tags: [users, mitre_remote_access_tools]

- rule: Terminal shell in container
  desc: A shell was used as the entrypoint/exec point into a container with an attached terminal.
  condition: >
    spawned_process and container
    and shell_procs and proc.tty != 0
    and container_entrypoint
  output: >
    A shell was spawned in a container with an attached terminal (user=%user.name %container.info
    shell=%proc.name parent=%proc.pname cmdline=%proc.cmdline terminal=%proc.tty container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [container, shell, mitre_execution]

# For some container types (mesos), there isn't a container image to
# work with, and the container name is autogenerated, so there isn't
# any stable aspect of the software to work with. In this case, we
# fall back to allowing certain command lines.

- list: known_shell_spawn_cmdlines
  items: [
    '"sh -c uname -p 2> /dev/null"',
    '"sh -c uname -s 2>&1"',
    '"sh -c uname -r 2>&1"',
    '"sh -c uname -v 2>&1"',
    '"sh -c uname -a 2>&1"',
    '"sh -c ruby -v 2>&1"',
    '"sh -c getconf CLK_TCK"',
    '"sh -c getconf PAGESIZE"',
    '"sh -c LC_ALL=C LANG=C /sbin/ldconfig -p 2>/dev/null"',
    '"sh -c LANG=C /sbin/ldconfig -p 2>/dev/null"',
    '"sh -c /sbin/ldconfig -p 2>/dev/null"',
    '"sh -c stty -a 2>/dev/null"',
    '"sh -c stty -a < /dev/tty"',
    '"sh -c stty -g < /dev/tty"',
    '"sh -c node index.js"',
    '"sh -c node index"',
    '"sh -c node ./src/start.js"',
    '"sh -c node app.js"',
    '"sh -c node -e \"require(''nan'')\""',
    '"sh -c node -e \"require(''nan'')\")"',
    '"sh -c node $NODE_DEBUG_OPTION index.js "',
    '"sh -c crontab -l 2"',
    '"sh -c lsb_release -a"',
    '"sh -c lsb_release -is 2>/dev/null"',
    '"sh -c whoami"',
    '"sh -c node_modules/.bin/bower-installer"',
    '"sh -c /bin/hostname -f 2> /dev/null"',
    '"sh -c locale -a"',
    '"sh -c  -t -i"',
    '"sh -c openssl version"',
    '"bash -c id -Gn kafadmin"',
    '"sh -c /bin/sh -c ''date +%%s''"'
    ]

# This list allows for easy additions to the set of commands allowed
# to run shells in containers without having to without having to copy
# and override the entire run shell in container macro. Once
# https://github.com/draios/falco/issues/255 is fixed this will be a
# bit easier, as someone could append of any of the existing lists.
- list: user_known_shell_spawn_binaries
  items: []

# This macro allows for easy additions to the set of commands allowed
# to run shells in containers without having to override the entire
# rule. Its default value is an expression that always is false, which
# becomes true when the "not ..." in the rule is applied.
- macro: user_shell_container_exclusions
  condition: (never_true)

- macro: login_doing_dns_lookup
  condition: (proc.name=login and fd.l4proto=udp and fd.sport=53)

# sockfamily ip is to exclude certain processes (like 'groups') that communicate on unix-domain sockets
# systemd can listen on ports to launch things like sshd on demand
- rule: System procs network activity
  desc: any network activity performed by system binaries that are not expected to send or receive any network traffic
  condition: >
    (fd.sockfamily = ip and (system_procs or proc.name in (shell_binaries)))
    and (inbound_outbound)
    and not proc.name in (systemd, hostid, id)
    and not login_doing_dns_lookup
  output: >
    Known system binary sent/received network traffic
    (user=%user.name command=%proc.cmdline connection=%fd.name container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [network, mitre_exfiltration]

# When filled in, this should look something like:
# (proc.env contains "HTTP_PROXY=http://my.http.proxy.com ")
# The trailing space is intentional so avoid matching on prefixes of
# the actual proxy.
- macro: allowed_ssh_proxy_env
  condition: (always_true)

- list: http_proxy_binaries
  items: [curl, wget]

- macro: http_proxy_procs
  condition: (proc.name in (http_proxy_binaries))

- rule: Program run with disallowed http proxy env
  desc: An attempt to run a program with a disallowed HTTP_PROXY environment variable
  condition: >
    spawned_process and
    http_proxy_procs and
    not allowed_ssh_proxy_env and
    proc.env icontains HTTP_PROXY
  output: >
    Program run with disallowed HTTP_PROXY environment variable
    (user=%user.name command=%proc.cmdline env=%proc.env parent=%proc.pname container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [host, users]

# In some environments, any attempt by a interpreted program (perl,
# python, ruby, etc) to listen for incoming connections or perform
# outgoing connections might be suspicious. These rules are not
# enabled by default, but you can modify the following macros to
# enable them.

- macro: consider_interpreted_inbound
  condition: (never_true)

- macro: consider_interpreted_outbound
  condition: (never_true)

- rule: Interpreted procs inbound network activity
  desc: Any inbound network activity performed by any interpreted program (perl, python, ruby, etc.)
  condition: >
    (inbound and consider_interpreted_inbound
     and interpreted_procs)
  output: >
    Interpreted program received/listened for network traffic
    (user=%user.name command=%proc.cmdline connection=%fd.name container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [network, mitre_exfiltration]

- rule: Interpreted procs outbound network activity
  desc: Any outbound network activity performed by any interpreted program (perl, python, ruby, etc.)
  condition: >
    (outbound and consider_interpreted_outbound
     and interpreted_procs)
  output: >
    Interpreted program performed outgoing network connection
    (user=%user.name command=%proc.cmdline connection=%fd.name container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [network, mitre_exfiltration]

- list: openvpn_udp_ports
  items: [1194, 1197, 1198, 8080, 9201]

- list: l2tp_udp_ports
  items: [500, 1701, 4500, 10000]

- list: statsd_ports
  items: [8125]

- list: ntp_ports
  items: [123]

# Some applications will connect a udp socket to an address only to
# test connectivity. Assuming the udp connect works, they will follow
# up with a tcp connect that actually sends/receives data.
#
# With that in mind, we listed a few commonly seen ports here to avoid
# some false positives. In addition, we make the main rule opt-in, so
# it's disabled by default.

- list: test_connect_ports
  items: [0, 9, 80, 3306]

- macro: do_unexpected_udp_check
  condition: (never_true)

- list: expected_udp_ports
  items: [53, openvpn_udp_ports, l2tp_udp_ports, statsd_ports, ntp_ports, test_connect_ports]

- macro: expected_udp_traffic
  condition: fd.port in (expected_udp_ports)

- rule: Unexpected UDP Traffic
  desc: UDP traffic not on port 53 (DNS) or other commonly used ports
  condition: (inbound_outbound) and do_unexpected_udp_check and fd.l4proto=udp and not expected_udp_traffic
  output: >
    Unexpected UDP Traffic Seen
    (user=%user.name command=%proc.cmdline connection=%fd.name proto=%fd.l4proto evt=%evt.type %evt.args container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [network, mitre_exfiltration]

# With the current restriction on system calls handled by falco
# (e.g. excluding read/write/sendto/recvfrom/etc, this rule won't
# trigger).
# - rule: Ssh error in syslog
#   desc: any ssh errors (failed logins, disconnects, ...) sent to syslog
#   condition: syslog and ssh_error_message and evt.dir = <
#   output: "sshd sent error message to syslog (error=%evt.buffer)"
#   priority: WARNING

- macro: somebody_becoming_themself
  condition: ((user.name=nobody and evt.arg.uid=nobody) or
              (user.name=www-data and evt.arg.uid=www-data) or
              (user.name=_apt and evt.arg.uid=_apt) or
              (user.name=postfix and evt.arg.uid=postfix) or
              (user.name=pki-agent and evt.arg.uid=pki-agent) or
              (user.name=pki-acme and evt.arg.uid=pki-acme) or
              (user.name=nfsnobody and evt.arg.uid=nfsnobody) or
              (user.name=postgres and evt.arg.uid=postgres))

- macro: nrpe_becoming_nagios
  condition: (proc.name=nrpe and evt.arg.uid=nagios)

# In containers, the user name might be for a uid that exists in the
# container but not on the host. (See
# https://github.com/draios/sysdig/issues/954). So in that case, allow
# a setuid.
- macro: known_user_in_container
  condition: (container and user.name != "N/A")

# Add conditions to this macro (probably in a separate file,
# overwriting this macro) to allow for specific combinations of
# programs changing users by calling setuid.
#
# In this file, it just takes one of the condition in the base macro
# and repeats it.
- macro: user_known_non_sudo_setuid_conditions
  condition: user.name=root

# sshd, mail programs attempt to setuid to root even when running as non-root. Excluding here to avoid meaningless FPs
- rule: Non sudo setuid
  desc: >
    an attempt to change users by calling setuid. sudo/su are excluded. users "root" and "nobody"
    suing to itself are also excluded, as setuid calls typically involve dropping privileges.
  condition: >
    evt.type=setuid and evt.dir=>
    and (known_user_in_container or not container)
    and not user.name=root
    and not somebody_becoming_themself
    and not proc.name in (known_setuid_binaries, userexec_binaries, mail_binaries, docker_binaries,
                          nomachine_binaries)
    and not proc.name startswith "runc:"
    and not java_running_sdjagent
    and not nrpe_becoming_nagios
    and not user_known_non_sudo_setuid_conditions
  output: >
    Unexpected setuid call by non-sudo, non-root program (user=%user.name cur_uid=%user.uid parent=%proc.pname
    command=%proc.cmdline uid=%evt.arg.uid container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [users, mitre_privilege_escalation]

- rule: User mgmt binaries
  desc: >
    activity by any programs that can manage users, passwords, or permissions. sudo and su are excluded.
    Activity in containers is also excluded--some containers create custom users on top
    of a base linux distribution at startup.
    Some innocuous commandlines that don't actually change anything are excluded.
  condition: >
    spawned_process and proc.name in (user_mgmt_binaries) and
    not proc.name in (su, sudo, lastlog, nologin, unix_chkpwd) and not container and
    not proc.pname in (cron_binaries, systemd, systemd.postins, udev.postinst, run-parts) and
    not proc.cmdline startswith "passwd -S" and
    not proc.cmdline startswith "useradd -D" and
    not proc.cmdline startswith "systemd --version" and
    not run_by_qualys and
    not run_by_sumologic_securefiles and
    not run_by_yum and
    not run_by_ms_oms and
    not run_by_google_accounts_daemon
  output: >
    User management binary command run outside of container
    (user=%user.name command=%proc.cmdline parent=%proc.pname gparent=%proc.aname[2] ggparent=%proc.aname[3] gggparent=%proc.aname[4])
  priority: NOTICE
  tags: [host, users, mitre_persistence]

- list: allowed_dev_files
  items: [
    /dev/null, /dev/stdin, /dev/stdout, /dev/stderr,
    /dev/random, /dev/urandom, /dev/console, /dev/kmsg
    ]

# (we may need to add additional checks against false positives, see:
# https://bugs.launchpad.net/ubuntu/+source/rkhunter/+bug/86153)
- rule: Create files below dev
  desc: creating any files below /dev other than known programs that manage devices. Some rootkits hide files in /dev.
  condition: >
    fd.directory = /dev and
    (evt.type = creat or (evt.type = open and evt.arg.flags contains O_CREAT))
    and not proc.name in (dev_creation_binaries)
    and not fd.name in (allowed_dev_files)
    and not fd.name startswith /dev/tty
  output: "File created below /dev by untrusted program (user=%user.name command=%proc.cmdline file=%fd.name container_id=%container.id image=%container.image.repository)"
  priority: ERROR
  tags: [filesystem, mitre_persistence]


# In a local/user rules file, you could override this macro to
# explicitly enumerate the container images that you want to allow
# access to EC2 metadata. In this main falco rules file, there isn't
# any way to know all the containers that should have access, so any
# container is alllowed, by repeating the "container" macro. In the
# overridden macro, the condition would look something like
# (container.image.repository = vendor/container-1 or
# container.image.repository = vendor/container-2 or ...)
- macro: ec2_metadata_containers
  condition: container

# On EC2 instances, 169.254.169.254 is a special IP used to fetch
# metadata about the instance. It may be desirable to prevent access
# to this IP from containers.
- rule: Contact EC2 Instance Metadata Service From Container
  desc: Detect attempts to contact the EC2 Instance Metadata Service from a container
  condition: outbound and fd.sip="169.254.169.254" and container and not ec2_metadata_containers
  output: Outbound connection to EC2 instance metadata service (command=%proc.cmdline connection=%fd.name %container.info image=%container.image.repository:%container.image.tag)
  priority: NOTICE
  tags: [network, aws, container, mitre_discovery]

# In a local/user rules file, you should override this macro with the
# IP address of your k8s api server. The IP 1.2.3.4 is a placeholder
# IP that is not likely to be seen in practice.
- macro: k8s_api_server
  condition: (fd.sip="1.2.3.4" and fd.sport=8080)

# In a local/user rules file, list the container images that are
# allowed to contact the K8s API Server from within a container. This
# might cover cases where the K8s infrastructure itself is running
# within a container.
- macro: k8s_containers
  condition: >
    (container.image.repository in (gcr.io/google_containers/hyperkube-amd64,
     gcr.io/google_containers/kube2sky, sysdig/agent, sysdig/falco,
     sysdig/sysdig))

- rule: Contact K8S API Server From Container
  desc: Detect attempts to contact the K8S API Server from a container
  condition: outbound and k8s_api_server and container and not k8s_containers
  output: Unexpected connection to K8s API Server from container (command=%proc.cmdline %container.info image=%container.image.repository:%container.image.tag connection=%fd.name)
  priority: NOTICE
  tags: [network, k8s, container, mitre_discovery]

# In a local/user rules file, list the container images that are
# allowed to contact NodePort services from within a container. This
# might cover cases where the K8s infrastructure itself is running
# within a container.
#
# By default, all containers are allowed to contact NodePort services.
- macro: nodeport_containers
  condition: container

- rule: Unexpected K8s NodePort Connection
  desc: Detect attempts to use K8s NodePorts from a container
  condition: (inbound_outbound) and fd.sport >= 30000 and fd.sport <= 32767 and container and not nodeport_containers
  output: Unexpected K8s NodePort Connection (command=%proc.cmdline connection=%fd.name container_id=%container.id image=%container.image.repository)
  priority: NOTICE
  tags: [network, k8s, container, mitre_port_knocking]

- list: network_tool_binaries
  items: [nc, ncat, nmap, dig, tcpdump, tshark, ngrep]

- macro: network_tool_procs
  condition: (proc.name in (network_tool_binaries))

# Container is supposed to be immutable. Package management should be done in building the image.
- rule: Launch Package Management Process in Container
  desc: Package management process ran inside container
  condition: >
    spawned_process and container and user.name != "_apt" and package_mgmt_procs and not package_mgmt_ancestor_procs
  output: >
    Package management process launched in container (user=%user.name
    command=%proc.cmdline container_id=%container.id container_name=%container.name image=%container.image.repository:%container.image.tag)
  priority: ERROR
  tags: [process, mitre_persistence]

- rule: Netcat Remote Code Execution in Container
  desc: Netcat Program runs inside container that allows remote code execution
  condition: >
    spawned_process and container and
    ((proc.name = "nc" and (proc.args contains "-e" or proc.args contains "-c")) or
     (proc.name = "ncat" and (proc.args contains "--sh-exec" or proc.args contains "--exec" or proc.args contains "-e "
                              or proc.args contains "-c " or proc.args contains "--lua-exec"))
    )
  output: >
    Netcat runs inside container that allows remote code execution (user=%user.name
    command=%proc.cmdline container_id=%container.id container_name=%container.name image=%container.image.repository:%container.image.tag)
  priority: WARNING
  tags: [network, process, mitre_execution]

- rule: Launch Suspicious Network Tool in Container
  desc: Detect network tools launched inside container
  condition: >
    spawned_process and container and network_tool_procs
  output: >
    Network tool launched in container (user=%user.name command=%proc.cmdline parent_process=%proc.pname
    container_id=%container.id container_name=%container.name image=%container.image.repository:%container.image.tag)
  priority: NOTICE
  tags: [network, process, mitre_discovery, mitre_exfiltration]

# This rule is not enabled by default, as there are legitimate use
# cases for these tools on hosts. If you want to enable it, modify the
# following macro.
- macro: consider_network_tools_on_host
  condition: (never_true)

- rule: Launch Suspicious Network Tool on Host
  desc: Detect network tools launched on the host
  condition: >
    spawned_process and
    not container and
    consider_network_tools_on_host and
    network_tool_procs
  output: >
    Network tool launched on host (user=%user.name command=%proc.cmdline parent_process=%proc.pname)
  priority: NOTICE
  tags: [network, process, mitre_discovery, mitre_exfiltration]

- list: grep_binaries
  items: [grep, egrep, fgrep]

- macro: grep_commands
  condition: (proc.name in (grep_binaries))

# a less restrictive search for things that might be passwords/ssh/user etc.
- macro: grep_more
  condition: (never_true)

- macro: private_key_or_password
  condition: >
    (proc.args icontains "BEGIN PRIVATE" or
     proc.args icontains "BEGIN RSA PRIVATE" or
     proc.args icontains "BEGIN DSA PRIVATE" or
     proc.args icontains "BEGIN EC PRIVATE" or
     (grep_more and
      (proc.args icontains " pass " or
       proc.args icontains " ssh " or
       proc.args icontains " user "))
    )

- rule: Search Private Keys or Passwords
  desc: >
    Detect grep private keys or passwords activity.
  condition: >
    (spawned_process and
     ((grep_commands and private_key_or_password) or
      (proc.name = "find" and (proc.args contains "id_rsa" or proc.args contains "id_dsa")))
    )
  output: >
    Grep private keys or passwords activities found
    (user=%user.name command=%proc.cmdline container_id=%container.id container_name=%container.name
    image=%container.image.repository:%container.image.tag)
  priority:
    WARNING
  tags: [process, mitre_credential_access]

- list: log_directories
  items: [/var/log, /dev/log]

- list: log_files
  items: [syslog, auth.log, secure, kern.log, cron, user.log, dpkg.log, last.log, yum.log, access_log, mysql.log, mysqld.log]

- macro: access_log_files
  condition: (fd.directory in (log_directories) or fd.filename in (log_files))

# a placeholder for whitelist log files that could be cleared. Recommend the macro as (fd.name startswith "/var/log/app1*")
- macro: allowed_clear_log_files
  condition: (never_true)

- macro: trusted_logging_images
  condition: (container.image.repository endswith "splunk/fluentd-hec")

- rule: Clear Log Activities
  desc: Detect clearing of critical log files
  condition: >
    open_write and
    access_log_files and
    evt.arg.flags contains "O_TRUNC" and
    not trusted_logging_images and
    not allowed_clear_log_files
  output: >
    Log files were tampered (user=%user.name command=%proc.cmdline file=%fd.name container_id=%container.id image=%container.image.repository)
  priority:
    WARNING
  tags: [file, mitre_defense_evasion]

- list: data_remove_commands
  items: [shred, mkfs, mke2fs]

- macro: clear_data_procs
  condition: (proc.name in (data_remove_commands))

- rule: Remove Bulk Data from Disk
  desc: Detect process running to clear bulk data from disk
  condition: spawned_process and clear_data_procs
  output: >
    Bulk data has been removed from disk (user=%user.name command=%proc.cmdline file=%fd.name container_id=%container.id image=%container.image.repository)
  priority:
    WARNING
  tags: [process, mitre_persistence]

- rule: Delete or rename shell history
  desc: Detect shell history deletion
  condition: >
    (modify and (
      evt.arg.name contains "bash_history" or
      evt.arg.name contains "zsh_history" or
      evt.arg.name contains "fish_read_history" or
      evt.arg.name endswith "fish_history" or
      evt.arg.oldpath contains "bash_history" or
      evt.arg.oldpath contains "zsh_history" or
      evt.arg.oldpath contains "fish_read_history" or
      evt.arg.oldpath endswith "fish_history" or
      evt.arg.path contains "bash_history" or
      evt.arg.path contains "zsh_history" or
      evt.arg.path contains "fish_read_history" or
      evt.arg.path endswith "fish_history")) or
    (open_write and (
      fd.name contains "bash_history" or
      fd.name contains "zsh_history" or
      fd.name contains "fish_read_history" or
      fd.name endswith "fish_history") and evt.arg.flags contains "O_TRUNC")
  output: >
    Shell history had been deleted or renamed (user=%user.name type=%evt.type command=%proc.cmdline fd.name=%fd.name name=%evt.arg.name path=%evt.arg.path oldpath=%evt.arg.oldpath %container.info)
  priority:
    WARNING
  tag: [process, mitre_defense_evation]

- macro: consider_all_chmods
  condition: (always_true)

- list: user_known_chmod_applications
  items: []

- rule: Set Setuid or Setgid bit
  desc: >
    When the setuid or setgid bits are set for an application,
    this means that the application will run with the privileges of the owning user or group respectively.
    Detect setuid or setgid bits set via chmod
  condition: consider_all_chmods and chmod and (evt.arg.mode contains "S_ISUID" or evt.arg.mode contains "S_ISGID") and not proc.cmdline in (user_known_chmod_applications)
  output: >
    Setuid or setgid bit is set via chmod (fd=%evt.arg.fd filename=%evt.arg.filename mode=%evt.arg.mode user=%user.name
    command=%proc.cmdline container_id=%container.id container_name=%container.name image=%container.image.repository:%container.image.tag)
  priority:
    NOTICE
  tag: [process, mitre_persistence]

- list: exclude_hidden_directories
  items: [/root/.cassandra]

# To use this rule, you should modify consider_hidden_file_creation.
- macro: consider_hidden_file_creation
  condition: (never_true)

- rule: Create Hidden Files or Directories
  desc: Detect hidden files or directories created
  condition: >
    (consider_hidden_file_creation and (
      (modify and evt.arg.newpath contains "/.") or
      (mkdir and evt.arg.path contains "/.") or
      (open_write and evt.arg.flags contains "O_CREAT" and fd.name contains "/." and not fd.name pmatch (exclude_hidden_directories)))
    )
  output: >
    Hidden file or directory created (user=%user.name command=%proc.cmdline
    file=%fd.name newpath=%evt.arg.newpath container_id=%container.id container_name=%container.name image=%container.image.repository:%container.image.tag)
  priority:
    NOTICE
  tag: [file, mitre_persistence]

- list: remote_file_copy_binaries
  items: [rsync, scp, sftp, dcp]

- macro: remote_file_copy_procs
  condition: (proc.name in (remote_File_copy_binaries))

- rule: Launch Remote File Copy Tools in Container
  desc: Detect remote file copy tools launched in container
  condition: >
    spawned_process and container and remote_file_copy_procs
  output: >
    Remote file copy tool launched in container (user=%user.name command=%proc.cmdline parent_process=%proc.pname
    container_id=%container.id container_name=%container.name image=%container.image.repository:%container.image.tag)
  priority: NOTICE
  tags: [network, process, mitre_lateral_movement, mitre_exfiltration]

- rule: Create Symlink Over Sensitive Files
  desc: Detect symlink created over sensitive files
  condition: >
    create_symlink and
    (evt.arg.target in (sensitive_file_names) or evt.arg.target in (sensitive_directory_names))
  output: >
    Symlinks created over senstivie files (user=%user.name command=%proc.cmdline target=%evt.arg.target linkpath=%evt.arg.linkpath parent_process=%proc.pname)
  priority: NOTICE
  tags: [file, mitre_exfiltration]

- list: miner_ports
  items: [
        25, 3333, 3334, 3335, 3336, 3357, 4444,
        5555, 5556, 5588, 5730, 6099, 6666, 7777,
        7778, 8000, 8001, 8008, 8080, 8118, 8333,
        8888, 8899, 9332, 9999, 14433, 14444,
        45560, 45700
    ]

- list: miner_domains
  items: [
      "asia1.ethpool.org","ca.minexmr.com",
      "cn.stratum.slushpool.com","de.minexmr.com",
      "eth-ar.dwarfpool.com","eth-asia.dwarfpool.com",
      "eth-asia1.nanopool.org","eth-au.dwarfpool.com",
      "eth-au1.nanopool.org","eth-br.dwarfpool.com",
      "eth-cn.dwarfpool.com","eth-cn2.dwarfpool.com",
      "eth-eu.dwarfpool.com","eth-eu1.nanopool.org",
      "eth-eu2.nanopool.org","eth-hk.dwarfpool.com",
      "eth-jp1.nanopool.org","eth-ru.dwarfpool.com",
      "eth-ru2.dwarfpool.com","eth-sg.dwarfpool.com",
      "eth-us-east1.nanopool.org","eth-us-west1.nanopool.org",
      "eth-us.dwarfpool.com","eth-us2.dwarfpool.com",
      "eu.stratum.slushpool.com","eu1.ethermine.org",
      "eu1.ethpool.org","fr.minexmr.com",
      "mine.moneropool.com","mine.xmrpool.net",
      "pool.minexmr.com","pool.monero.hashvault.pro",
      "pool.supportxmr.com","sg.minexmr.com",
      "sg.stratum.slushpool.com","stratum-eth.antpool.com",
      "stratum-ltc.antpool.com","stratum-zec.antpool.com",
      "stratum.antpool.com","us-east.stratum.slushpool.com",
      "us1.ethermine.org","us1.ethpool.org",
      "us2.ethermine.org","us2.ethpool.org",
      "xmr-asia1.nanopool.org","xmr-au1.nanopool.org",
      "xmr-eu1.nanopool.org","xmr-eu2.nanopool.org",
      "xmr-jp1.nanopool.org","xmr-us-east1.nanopool.org",
      "xmr-us-west1.nanopool.org","xmr.crypto-pool.fr",
      "xmr.pool.minergate.com"
      ]

- list: https_miner_domains
  items: [
    "ca.minexmr.com",
    "cn.stratum.slushpool.com",
    "de.minexmr.com",
    "fr.minexmr.com",
    "mine.moneropool.com",
    "mine.xmrpool.net",
    "pool.minexmr.com",
    "sg.minexmr.com",
    "stratum-eth.antpool.com",
    "stratum-ltc.antpool.com",
    "stratum-zec.antpool.com",
    "stratum.antpool.com",
    "xmr.crypto-pool.fr"
  ]

- list: http_miner_domains
  items: [
    "ca.minexmr.com",
    "de.minexmr.com",
    "fr.minexmr.com",
    "mine.moneropool.com",
    "mine.xmrpool.net",
    "pool.minexmr.com",
    "sg.minexmr.com",
    "xmr.crypto-pool.fr"
  ]

# Add rule based on crypto mining IOCs
- macro: minerpool_https
  condition: (fd.sport="443" and fd.sip.name in (https_miner_domains))

- macro: minerpool_http
  condition: (fd.sport="80" and fd.sip.name in (http_miner_domains))

- macro: minerpool_other
  condition: (fd.sport in (miner_ports) and fd.sip.name in (miner_domains))

- macro: net_miner_pool
  condition: (evt.type in (sendto, sendmsg) and evt.dir=< and ((minerpool_http) or (minerpool_https) or (minerpool_other)))

- rule: Detect outbound connections to common miner pool ports
  desc: Miners typically connect to miner pools on common ports.
  condition: net_miner_pool
  output: Outbound connection to IP/Port flagged by cryptoioc.ch (command=%proc.cmdline port=%fd.rport ip=%fd.rip container=%container.info image=%container.image.repository)
  priority: CRITICAL
  tags: [network, mitre_execution]

- rule: Detect crypto miners using the Stratum protocol
  desc: Miners typically specify the mining pool to connect to with a URI that begins with 'stratum+tcp'
  condition: spawned_process and proc.cmdline contains "stratum+tcp"
  output: Possible miner running (command=%proc.cmdline container=%container.info image=%container.image.repository)
  priority: CRITICAL
  tags: [process, mitre_execution]

# Application rules have moved to application_rules.yaml. Please look
# there if you want to enable them by adding to
# falco_rules.local.yaml.

`,
	"falco.yaml": `#
# Copyright (C) 2016-2018 Draios Inc dba Sysdig.
#
# This file is part of falco .
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# File(s) or Directories containing Falco rules, loaded at startup.
# The name "rules_file" is only for backwards compatibility.
# If the entry is a file, it will be read directly. If the entry is a directory,
# every file in that directory will be read, in alphabetical order.
#
# falco_rules.yaml ships with the falco package and is overridden with
# every new software version. falco_rules.local.yaml is only created
# if it doesn't exist. If you want to customize the set of rules, add
# your customizations to falco_rules.local.yaml.
#
# The files will be read in the order presented here, so make sure if
# you have overrides they appear in later files.
rules_file:
 - /etc/falco/falco_rules.yaml
 - /etc/falco/falco_rules.local.yaml
 - /etc/falco/k8s_audit_rules.yaml
 - /etc/falco/rules.d

# If true, the times displayed in log messages and output messages
# will be in ISO 8601. By default, times are displayed in the local
# time zone, as governed by /etc/localtime.
time_format_iso_8601: false

# Whether to output events in json or text
json_output: false

# When using json output, whether or not to include the "output" property
# itself (e.g. "File below a known binary directory opened for writing
# (user=root ....") in the json output.
json_include_output_property: true

# Send information logs to stderr and/or syslog Note these are *not* security
# notification logs! These are just Falco lifecycle (and possibly error) logs.
log_stderr: true
log_syslog: true

# Minimum log level to include in logs. Note: these levels are
# separate from the priority field of rules. This refers only to the
# log level of falco's internal logging. Can be one of "emergency",
# "alert", "critical", "error", "warning", "notice", "info", "debug".
log_level: info

# Minimum rule priority level to load and run. All rules having a
# priority more severe than this level will be loaded/run.  Can be one
# of "emergency", "alert", "critical", "error", "warning", "notice",
# "info", "debug".
priority: debug

# Whether or not output to any of the output channels below is
# buffered. Defaults to false
buffered_outputs: false

# Falco uses a shared buffer between the kernel and userspace to pass
# system call information. When falco detects that this buffer is
# full and system calls have been dropped, it can take one or more of
# the following actions:
#   - "ignore": do nothing. If an empty list is provided, ignore is assumed.
#   - "log": log a CRITICAL message noting that the buffer was full.
#   - "alert": emit a falco alert noting that the buffer was full.
#   - "exit": exit falco with a non-zero rc.
#
# The rate at which log/alert messages are emitted is governed by a
# token bucket. The rate corresponds to one message every 30 seconds
# with a burst of 10 messages.

syscall_event_drops:
  actions:
    - log
    - alert
  rate: .03333
  max_burst: 10

# A throttling mechanism implemented as a token bucket limits the
# rate of falco notifications. This throttling is controlled by the following configuration
# options:
#  - rate: the number of tokens (i.e. right to send a notification)
#    gained per second. Defaults to 1.
#  - max_burst: the maximum number of tokens outstanding. Defaults to 1000.
#
# With these defaults, falco could send up to 1000 notifications after
# an initial quiet period, and then up to 1 notification per second
# afterward. It would gain the full burst back after 1000 seconds of
# no activity.

outputs:
  rate: 1
  max_burst: 1000

# Where security notifications should go.
# Multiple outputs can be enabled.

syslog_output:
  enabled: true

# If keep_alive is set to true, the file will be opened once and
# continuously written to, with each output message on its own
# line. If keep_alive is set to false, the file will be re-opened
# for each output message.
#
# Also, the file will be closed and reopened if falco is signaled with
# SIGUSR1.

file_output:
  enabled: false
  keep_alive: false
  filename: ./events.txt

stdout_output:
  enabled: true

# Falco contains an embedded webserver that can be used to accept K8s
# Audit Events. These config options control the behavior of that
# webserver. (By default, the webserver is disabled).
#
# The ssl_certificate is a combination SSL Certificate and corresponding
# key contained in a single file. You can generate a key/cert as follows:
#
# $ openssl req -newkey rsa:2048 -nodes -keyout key.pem -x509 -days 365 -out certificate.pem
# $ cat certificate.pem key.pem > falco.pem
# $ sudo cp falco.pem /etc/falco/falco.pem

webserver:
  enabled: true
  listen_port: 8765
  k8s_audit_endpoint: /k8s_audit
  ssl_enabled: false
  ssl_certificate: /etc/falco/falco.pem

# Possible additional things you might want to do with program output:
#   - send to a slack webhook:
#         program: "jq '{text: .output}' | curl -d @- -X POST https://hooks.slack.com/services/XXX"
#   - logging (alternate method than syslog):
#         program: logger -t falco-test
#   - send over a network connection:
#         program: nc host.example.com 80

# If keep_alive is set to true, the program will be started once and
# continuously written to, with each output message on its own
# line. If keep_alive is set to false, the program will be re-spawned
# for each output message.
#
# Also, the program will be closed and reopened if falco is signaled with
# SIGUSR1.
program_output:
  enabled: false
  keep_alive: false
  program: "jq '{text: .output}' | curl -d @- -X POST https://hooks.slack.com/services/XXX"

http_output:
  enabled: false
  url: http://some.url`,
	"k8s_audit_rules.yaml": `#
# Copyright (C) 2016-2018 Draios Inc dba Sysdig.
#
# This file is part of falco.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
- required_engine_version: 2

# Like always_true/always_false, but works with k8s audit events
- macro: k8s_audit_always_true
  condition: (jevt.rawtime exists)

- macro: k8s_audit_never_true
  condition: (jevt.rawtime=0)

# Generally only consider audit events once the response has completed
- list: k8s_audit_stages
  items: ["ResponseComplete"]

# Generally exclude users starting with "system:"
- macro: non_system_user
  condition: (not ka.user.name startswith "system:")

# This macro selects the set of Audit Events used by the below rules.
- macro: kevt
  condition: (jevt.value[/stage] in (k8s_audit_stages))

- macro: kevt_started
  condition: (jevt.value[/stage]=ResponseStarted)

# If you wish to restrict activity to a specific set of users, override/append to this list.
- list: allowed_k8s_users
  items: ["minikube", "minikube-user", "kubelet", "kops"]

- rule: Disallowed K8s User
  desc: Detect any k8s operation by users outside of an allowed set of users.
  condition: kevt and non_system_user and not ka.user.name in (allowed_k8s_users)
  output: K8s Operation performed by user not in allowed list of users (user=%ka.user.name target=%ka.target.name/%ka.target.resource verb=%ka.verb uri=%ka.uri resp=%ka.response.code)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

# In a local/user rules file, you could override this macro to
# explicitly enumerate the container images that you want to run in
# your environment. In this main falco rules file, there isn't any way
# to know all the containers that can run, so any container is
# allowed, by using the always_true macro. In the overridden macro, the condition
# would look something like (ka.req.container.image.repository=my-repo/my-image)
- macro: allowed_k8s_containers
  condition: (k8s_audit_always_true)

- macro: response_successful
  condition: (ka.response.code startswith 2)

- macro: kcreate
  condition: ka.verb=create

- macro: kmodify
  condition: (ka.verb in (create,update,patch))

- macro: kdelete
  condition: ka.verb=delete

- macro: pod
  condition: ka.target.resource=pods and not ka.target.subresource exists

- macro: pod_subresource
  condition: ka.target.resource=pods and ka.target.subresource exists

- macro: deployment
  condition: ka.target.resource=deployments

- macro: service
  condition: ka.target.resource=services

- macro: configmap
  condition: ka.target.resource=configmaps

- macro: namespace
  condition: ka.target.resource=namespaces

- macro: serviceaccount
  condition: ka.target.resource=serviceaccounts

- macro: clusterrole
  condition: ka.target.resource=clusterroles

- macro: clusterrolebinding
  condition: ka.target.resource=clusterrolebindings

- macro: role
  condition: ka.target.resource=roles

- macro: health_endpoint
  condition: ka.uri=/healthz

- rule: Create Disallowed Pod
  desc: >
    Detect an attempt to start a pod with a container image outside of a list of allowed images.
  condition: kevt and pod and kcreate and not allowed_k8s_containers
  output: Pod started with container not in allowed list (user=%ka.user.name pod=%ka.resp.name ns=%ka.target.namespace image=%ka.req.container.image)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

- rule: Create Privileged Pod
  desc: >
    Detect an attempt to start a pod with a privileged container
  condition: kevt and pod and kcreate and ka.req.container.privileged=true and not ka.req.container.image.repository in (falco_privileged_images)
  output: Pod started with privileged container (user=%ka.user.name pod=%ka.resp.name ns=%ka.target.namespace image=%ka.req.container.image)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

- macro: sensitive_vol_mount
  condition: >
    (ka.req.volume.hostpath[/proc*]=true or
     ka.req.volume.hostpath[/var/run/docker.sock]=true or
     ka.req.volume.hostpath[/]=true or
     ka.req.volume.hostpath[/etc]=true or
     ka.req.volume.hostpath[/root*]=true)

- rule: Create Sensitive Mount Pod
  desc: >
    Detect an attempt to start a pod with a volume from a sensitive host directory (i.e. /proc).
    Exceptions are made for known trusted images.
  condition: kevt and pod and kcreate and sensitive_vol_mount and not ka.req.container.image.repository in (falco_sensitive_mount_images)
  output: Pod started with sensitive mount (user=%ka.user.name pod=%ka.resp.name ns=%ka.target.namespace image=%ka.req.container.image mounts=%jevt.value[/requestObject/spec/volumes])
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

# Corresponds to K8s CIS Benchmark 1.7.4
- rule: Create HostNetwork Pod
  desc: Detect an attempt to start a pod using the host network.
  condition: kevt and pod and kcreate and ka.req.container.host_network=true and not ka.req.container.image.repository in (falco_hostnetwork_images)
  output: Pod started using host network (user=%ka.user.name pod=%ka.resp.name ns=%ka.target.namespace image=%ka.req.container.image)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

- rule: Create NodePort Service
  desc: >
    Detect an attempt to start a service with a NodePort service type
  condition: kevt and service and kcreate and ka.req.service.type=NodePort
  output: NodePort Service Created (user=%ka.user.name service=%ka.target.name ns=%ka.target.namespace ports=%ka.req.service.ports)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

- macro: contains_private_credentials
  condition: >
    (ka.req.configmap.obj contains "aws_access_key_id" or
     ka.req.configmap.obj contains "aws-access-key-id" or
     ka.req.configmap.obj contains "aws_s3_access_key_id" or
     ka.req.configmap.obj contains "aws-s3-access-key-id" or
     ka.req.configmap.obj contains "password" or
     ka.req.configmap.obj contains "passphrase")

- rule: Create/Modify Configmap With Private Credentials
  desc: >
     Detect creating/modifying a configmap containing a private credential (aws key, password, etc.)
  condition: kevt and configmap and kmodify and contains_private_credentials
  output: K8s configmap with private credential (user=%ka.user.name verb=%ka.verb configmap=%ka.req.configmap.name config=%ka.req.configmap.obj)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

# Corresponds to K8s CIS Benchmark, 1.1.1.
- rule: Anonymous Request Allowed
  desc: >
    Detect any request made by the anonymous user that was allowed
  condition: kevt and ka.user.name=system:anonymous and ka.auth.decision!=reject and not health_endpoint
  output: Request by anonymous user allowed (user=%ka.user.name verb=%ka.verb uri=%ka.uri reason=%ka.auth.reason))
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

# Roughly corresponds to K8s CIS Benchmark, 1.1.12. In this case,
# notifies an attempt to exec/attach to a privileged container.

# Ideally, we'd add a more stringent rule that detects attaches/execs
# to a privileged pod, but that requires the engine for k8s audit
# events to be stateful, so it could know if a container named in an
# attach request was created privileged or not. For now, we have a
# less severe rule that detects attaches/execs to any pod.

- rule: Attach/Exec Pod
  desc: >
    Detect any attempt to attach/exec to a pod
  condition: kevt_started and pod_subresource and kcreate and ka.target.subresource in (exec,attach)
  output: Attach/Exec to pod (user=%ka.user.name pod=%ka.target.name ns=%ka.target.namespace action=%ka.target.subresource command=%ka.uri.param[command])
  priority: NOTICE
  source: k8s_audit
  tags: [k8s]

# In a local/user rules fie, you can append to this list to add additional allowed namespaces
- list: allowed_namespaces
  items: [kube-system, kube-public, default]

- rule: Create Disallowed Namespace
  desc: Detect any attempt to create a namespace outside of a set of known namespaces
  condition: kevt and namespace and kcreate and not ka.target.name in (allowed_namespaces)
  output: Disallowed namespace created (user=%ka.user.name ns=%ka.target.name)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

# Detect any new pod created in the kube-system namespace
- rule: Pod Created in Kube Namespace
  desc: Detect any attempt to create a pod in the kube-system or kube-public namespaces
  condition: kevt and pod and kcreate and ka.target.namespace in (kube-system, kube-public)
  output: Pod created in kube namespace (user=%ka.user.name pod=%ka.resp.name ns=%ka.target.namespace image=%ka.req.container.image)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

# Detect creating a service account in the kube-system/kube-public namespace
- rule: Service Account Created in Kube Namespace
  desc: Detect any attempt to create a serviceaccount in the kube-system or kube-public namespaces
  condition: kevt and serviceaccount and kcreate and ka.target.namespace in (kube-system, kube-public)
  output: Service account created in kube namespace (user=%ka.user.name serviceaccount=%ka.target.name ns=%ka.target.namespace)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

# Detect any modify/delete to any ClusterRole starting with
# "system:". "system:coredns" is excluded as changes are expected in
# normal operation.
- rule: System ClusterRole Modified/Deleted
  desc: Detect any attempt to modify/delete a ClusterRole/Role starting with system
  condition: kevt and (role or clusterrole) and (kmodify or kdelete) and (ka.target.name startswith "system:") and ka.target.name!="system:coredns"
  output: System ClusterRole/Role modified or deleted (user=%ka.user.name role=%ka.target.name ns=%ka.target.namespace action=%ka.verb)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

# Detect any attempt to create a ClusterRoleBinding to the cluster-admin user
# (exapand this to any built-in cluster role that does "sensitive" things)
- rule: Attach to cluster-admin Role
  desc: Detect any attempt to create a ClusterRoleBinding to the cluster-admin user
  condition: kevt and clusterrolebinding and kcreate and ka.req.binding.role=cluster-admin
  output: Cluster Role Binding to cluster-admin role (user=%ka.user.name subject=%ka.req.binding.subjects)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

- rule: ClusterRole With Wildcard Created
  desc: Detect any attempt to create a Role/ClusterRole with wildcard resources or verbs
  condition: kevt and (role or clusterrole) and kcreate and (ka.req.role.rules.resources contains '"*"' or ka.req.role.rules.verbs contains '"*"')
  output: Created Role/ClusterRole with wildcard (user=%ka.user.name role=%ka.target.name rules=%ka.req.role.rules)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

- macro: writable_verbs
  condition: >
    (ka.req.role.rules.verbs contains create or
     ka.req.role.rules.verbs contains update or
     ka.req.role.rules.verbs contains patch or
     ka.req.role.rules.verbs contains delete or
     ka.req.role.rules.verbs contains deletecollection)

- rule: ClusterRole With Write Privileges Created
  desc: Detect any attempt to create a Role/ClusterRole that can perform write-related actions
  condition: kevt and (role or clusterrole) and kcreate and writable_verbs
  output: Created Role/ClusterRole with write privileges (user=%ka.user.name role=%ka.target.name rules=%ka.req.role.rules)
  priority: NOTICE
  source: k8s_audit
  tags: [k8s]

- rule: ClusterRole With Pod Exec Created
  desc: Detect any attempt to create a Role/ClusterRole that can exec to pods
  condition: kevt and (role or clusterrole) and kcreate and ka.req.role.rules.resources contains "pods/exec"
  output: Created Role/ClusterRole with pod exec privileges (user=%ka.user.name role=%ka.target.name rules=%ka.req.role.rules)
  priority: WARNING
  source: k8s_audit
  tags: [k8s]

# The rules below this point are less discriminatory and generally
# represent a stream of activity for a cluster. If you wish to disable
# these events, modify the following macro.
- macro: consider_activity_events
  condition: (k8s_audit_always_true)

- macro: kactivity
  condition: (kevt and consider_activity_events)

- rule: K8s Deployment Created
  desc: Detect any attempt to create a deployment
  condition: (kactivity and kcreate and deployment and response_successful)
  output: K8s Deployment Created (user=%ka.user.name deployment=%ka.target.name ns=%ka.target.namespace resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s Deployment Deleted
  desc: Detect any attempt to delete a deployment
  condition: (kactivity and kdelete and deployment and response_successful)
  output: K8s Deployment Deleted (user=%ka.user.name deployment=%ka.target.name ns=%ka.target.namespace resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s Service Created
  desc: Detect any attempt to create a service
  condition: (kactivity and kcreate and service and response_successful)
  output: K8s Service Created (user=%ka.user.name service=%ka.target.name ns=%ka.target.namespace resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s Service Deleted
  desc: Detect any attempt to delete a service
  condition: (kactivity and kdelete and service and response_successful)
  output: K8s Service Deleted (user=%ka.user.name service=%ka.target.name ns=%ka.target.namespace resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s ConfigMap Created
  desc: Detect any attempt to create a configmap
  condition: (kactivity and kcreate and configmap and response_successful)
  output: K8s ConfigMap Created (user=%ka.user.name configmap=%ka.target.name ns=%ka.target.namespace resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s ConfigMap Deleted
  desc: Detect any attempt to delete a configmap
  condition: (kactivity and kdelete and configmap and response_successful)
  output: K8s ConfigMap Deleted (user=%ka.user.name configmap=%ka.target.name ns=%ka.target.namespace resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s Namespace Created
  desc: Detect any attempt to create a namespace
  condition: (kactivity and kcreate and namespace and response_successful)
  output: K8s Namespace Created (user=%ka.user.name namespace=%ka.target.name resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s Namespace Deleted
  desc: Detect any attempt to delete a namespace
  condition: (kactivity and non_system_user and kdelete and namespace and response_successful)
  output: K8s Namespace Deleted (user=%ka.user.name namespace=%ka.target.name resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s Serviceaccount Created
  desc: Detect any attempt to create a service account
  condition: (kactivity and kcreate and serviceaccount and response_successful)
  output: K8s Serviceaccount Created (user=%ka.user.name user=%ka.target.name ns=%ka.target.namespace resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s Serviceaccount Deleted
  desc: Detect any attempt to delete a service account
  condition: (kactivity and kdelete and serviceaccount and response_successful)
  output: K8s Serviceaccount Deleted (user=%ka.user.name user=%ka.target.name ns=%ka.target.namespace resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s Role/Clusterrole Created
  desc: Detect any attempt to create a cluster role/role
  condition: (kactivity and kcreate and (clusterrole or role) and response_successful)
  output: K8s Cluster Role Created (user=%ka.user.name role=%ka.target.name rules=%ka.req.role.rules resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s Role/Clusterrole Deleted
  desc: Detect any attempt to delete a cluster role/role
  condition: (kactivity and kdelete and (clusterrole or role) and response_successful)
  output: K8s Cluster Role Deleted (user=%ka.user.name role=%ka.target.name resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s Role/Clusterrolebinding Created
  desc: Detect any attempt to create a clusterrolebinding
  condition: (kactivity and kcreate and clusterrolebinding and response_successful)
  output: K8s Cluster Role Binding Created (user=%ka.user.name binding=%ka.target.name subjects=%ka.req.binding.subjects role=%ka.req.binding.role resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason foo=%ka.req.binding.subject.has_name[cluster-admin])
  priority: INFO
  source: k8s_audit
  tags: [k8s]

- rule: K8s Role/Clusterrolebinding Deleted
  desc: Detect any attempt to delete a clusterrolebinding
  condition: (kactivity and kdelete and clusterrolebinding and response_successful)
  output: K8s Cluster Role Binding Deleted (user=%ka.user.name binding=%ka.target.name resp=%ka.response.code decision=%ka.auth.decision reason=%ka.auth.reason)
  priority: INFO
  source: k8s_audit
  tags: [k8s]

# This rule generally matches all events, and as a result is disabled
# by default. If you wish to enable these events, modify the
# following macro.
#  condition: (jevt.rawtime exists)
- macro: consider_all_events
  condition: (k8s_audit_never_true)

- macro: kall
  condition: (kevt and consider_all_events)

- rule: All K8s Audit Events
  desc: Match all K8s Audit Events
  condition: kall
  output: K8s Audit Event received (user=%ka.user.name verb=%ka.verb uri=%ka.uri obj=%jevt.obj)
  priority: DEBUG
  source: k8s_audit
  tags: [k8s]
`,
}
