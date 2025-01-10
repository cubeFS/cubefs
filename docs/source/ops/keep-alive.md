# keep-alive script
The keepalive script will be available in version 3.5.1.

## Running Guide
Enter the/tool/keepalive directory, execute the installation script.
``` 
./cbfs_auto_install.sh arg1 arg2 arg3 arg4 arg5 arg6 arg7 arg8
```

* Before executing commands, users and volumes need to be created first
* Execute without root privileges, you need to add `user_allow_other` to `/etc/fuse.conf` before executing the command

## Parameter description
arg1: script placement directory, such as `/home/service/app/cfs'`
arg2: mountPoint The local directory to be mounted (if multiple instances need to be launched, they need to be distinguished)
arg3: volName The volume name applied for, such as ltptest
arg4: owner The account applied for, such as prod_xxx
arg5: accessKey
arg6: secretKey
arg7: master address
arg8: Client log directory (if multiple instances need to be launched, they need to be distinguished)
arg9: Read-only mount, optional parameter, true read-only mount.

## Running instructions
`cbfs_auto_install.sh` will automatically install `cbfs_restart.sh` and `cbfs_clean.sh` to the specified path `arg1`, create 'client.conf' in the 'arg1/arg3' path and copy cfs-client to that path, add the restart rule to `crontab`.By default, the cbfs_restart script is executed once per minute.

cbfs_restart.sh: used to restart the crashed client, automatically mount at a scheduled time
cbfs_clean.sh: used to clean up the mount point and crontab rules, needs to be executed manually