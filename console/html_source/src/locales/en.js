// import enLocale from '@jd/dui/dist/locale/en-US'
import enLocale from "element-ui/lib/locale/lang/en";

export default {
  languages: [
    {
      text: "Chinese",
      val: "zh"
    },
    {
      text: "English",
      val: "en"
    }
  ],
  chubaoFS: {
    nav: {
      Overview: "Overview",
      Servers: "Servers",
      Health: "Health",
      Volume: "Volume",
      Operations: "Operations",
      Alarm: "Alarm",
      Authorization: "Authorization",
      Cluster: "Cluster management",
      ClusterOverview: "Cluster Overview",
      ClusterDetail: "Cluster Detail"
    },
    login: {
      welcome: "Welcome！",
      subWelcome: "Sign in to access your account.",
      seeMore: "SEE MORE",
      subSeeMore: "Complete the form below to log in",
      name: "NAME",
      password: "PASSWORD",
      login: "LOG IN",
      SignOut: "Sign out"
    },
    overview: {
      cluster: {
        name: "Cluster",
        serverCount: "Server Count",
        dataPartitionCount: "DataNode Partition Count",
        leaderAddr: "Leader Address",
        metaPartitionCount: "MetaNode Partition Count",
        disableAutoAlloc: "Automatic space allocation",
        volumeCount: "Volume Count"
      },
      DataSize: "Data Size",
      VolumeCount: "Volume Count",
      Users: "Users",
      Details: "Detail",
      DataNode: "DataNode",
      MetaNode: "MetaNode",
      Top10Users: "Top 10 Users",
      TotalCapacity: "Total Capacity:",
      Size: "Size:",
      Available: "Available",
      Used: "Used",
      TotalVolumeCapacity: "Total volume capacity",
      UsedData: "Used data"
    },
    alarm: {
      name: "Alarm Info",
      detail: {
        ClusterID: "Cluster ID",
        Alarm: "Alarm",
        Host: "Host",
        Time: "Time"
      }
    },
    authorization: {
      AddUser: "Add User",
      detail: {
        UserName: "User Name",
        UserType: "User Type",
        Comments: "Comments",
        Actions: "Actions"
      },
      actions: {
        AddNewUser: "Add New User",
        UserName: "User Name",
        SetPassword: "Set Password",
        ConfirmPassword: "Confirm Password",
        Type: "Type",
        Comments: "Comments",
        NewPassword: "New Password",
        ConfirmNewPassword: "Confirm New Password",
        EditUserAuthorization: "Edit User Authorization"
      },
      tips: {
        delTip: "Are you sure you want to delete the User?",
        SetPasswordMsg: "Please enter Password",
        checkPasswordMsg:
          "Consists of letters and numbers. No less than 6 characters.No more than 30 characters.",
        checkUserNameMsg:
          "Consists of letters, numbers and underscores, and must start with a letter. No more than 20 characters.",
        confirmPasswordMsg1: "Please enter Confirmation password",
        confirmPasswordMsg2: "Confirmation password is not identical.",
        typeMsg: "Please select Type"
      }
    },
    operations: {
      name: "Operations",
      ServerManagement: {
        title: "Server Management",
        Cluster: "Cluster",
        IDC: "IDC",
        ServerIP: "Server IP",
        ServiceType: "Service Type",
        Status: "Status",
        IsWritable: "IsWritable",
        UsedCapacity: "Used Capacity",
        TotalCapacity: "Total Capacity",
        Used: "Used %",
        OffLineDisk: "Off-line Disk",
        OffLineServer: "Off-line Server",
        OffDistTip: "Please select the disk(s) to want to take-off:",
        offServerTip: "Are you sure to take off Server",
        diskTip: "Please select the disk(s) to want to take-off:",
        NoDisk: "No disk!",
        TakeOffServerTip: "Are you sure to take off Server"
      },
      VolumeManagement: {
        title: "Volume Management",
        CreateVolume: "Create Volume",
        VolumeName: "Volume Name",
        VolumeNumber: "Volume number",
        Name: "Name",
        TotalCapacity: "Total capacity",
        Used: "Used",
        UsedRate: "Used%",
        InodeCount: "Inode Count",
        Replications: "Replications",
        Status: "Status",
        TotalDataPartition: "Total Data Partition",
        OwnerID: "Owner ID",
        ZoneName: "Zone Name",
        AvailableDataPartition: "Available Data Partition",
        TotalMetapartition: "Total Metapartition",
        CreateTime: "Create time",
        IncreaseVolumeCapacity: "Increase Volume Capacity",
        CurrentCapacity: "Current capacity",
        Extendto: "Extend to",
        PermissionofVolume_Spark: "Permission of Volume_Spark",
        GrantPerimssion: "Grant Perimssion",
        Access: "Access",
        EditVolume: "Edit Volume",
        EditUserAccess: "Edit User Access",
        dpReplicaNumMsg: "Please select Replications",
        ownerMsg: "Please enter Owner ID",
        notesMsg: "Please enter Notes",
        userNameMsg: "Please enter User Name",
        accessMsg: "Please select Access",
        updateVolumeError:
          "don't support new replicaNum[3] larger than old dpReplicaNum[2]",
        extendVolumeError:
          "capacity[capacityVal] less than old capacity[oldCapacityVal]",
        deleteVolTip: "Are you sure to delete Volume",
        deletePermTip: "Are you sure you want to remove the user's permission?",
        checkVolNameTip:
          "The volume name can contain lowercase letters, numbers, and hyphens. It must start and end with lowercase letters or numbers, and be in the 3-63 length range.",
        checkOwnerIdTip:
          "Consists of letters, numbers and underscores, and must start with a letter. No more than 20 characters.",
        GrantPermissionto: "Grant Permission to"
      },
      MetaPartitionManagement: {
        title: "MetaPartition Management",
        PartitionID: "ID",
        dentryCount: "Dentry Count",
        inodeCount: "Inode Count",
        start: "Start",
        end: "End",
        isRecover: "Is Recover",
        missNodes: "Miss Nodes",
        replicaNum: "Replica num",
        status: "status",
        Hosts: "Hosts",
        VolName: "Volume Name",
        AddReplica: "Add Replica",
        DeleteReplica: "Delete Replica",
        DecommissionReplica: "Decommission Replica"
      },
      DataPartitionManagement: {
        title: "DataPartition Management",
        PartitionID: "ID",
        MissNodes: "Miss Nodes",
        ReplicaNum: "Replica num",
        Status: "status",
        LoadedTime: "Loaded Time",
        Hosts: "Hosts",
        VolName: "Volume Name",
        AddReplica: "Add Replica",
        DeleteReplica: "Delete Replica",
        DecommissionReplica: "Decommission Replica"
      }
    },
    servers: {
      Cluster: "Cluster",
      TotalVolume: "Total Volume",
      ServerList: "Server List",
      Master: "Master",
      Matanode: "Matanode",
      Datanode: "Datanode",
      Address: "Address",
      Leader: "Leader",
      Zone: "Zone",
      Used: "Used",
      PartitionCount: "Partition Count",
      Total: "Total",
      Available: "Available",
      UsedRate: "Used%",
      ReportTime: "Report Time",
      IsActive: "Is Active",
      identity: "ID",
      PartitionList: "Partition List",
      PartitionID: "Partition ID",
      Volume: "Volume",
      DiskPath: "DiskPath",
      ExtentCount: "Extent Count",
      NeedCompare: "Need Compare",
      Start: "Start",
      End: "End",
      Status: "Status"
    },
    userList: {
      UserList: "User list",
      User: "User",
      Data: "Data",
      VolumeCount: "Volume Count",
      DataPartitionCount: "Data Partition Count",
      MetaPartitionCount: "Meta Partition Count"
    },
    depot: {
      VolumeName: "Volume Name",
      Type: "Type",
      DeleteTime: "Delete Time",
      DeleteIp: "Delete IP",
      File: "File",
      Directory: "Directory"
    },
    cluster: {
      ClusterName: "Cluster Name",
      VolumeCount: "Volume quantity(individual)",
      StorageCapacity: "Storage Capacity(TB)",
      UsageCapacity: "Usage Capacity(TB)",
      UsageRate: "Usage Rate",
      YesterdayAdd: "New usage capacity added yesterday(TB)",
      ClusterHealth: "Cluster Health"
    },
    DepotManagement: {
      ClearAll: "Are you sure you want to empty the recycle bin?",
      Recover: "Are you sure you want to restore this file/directory?"
    },
    volume: {
      AccessKeys: "Access Keys",
      CreateVolume: "Create Volume",
      Volumenumber: "Volume Number",
      Info: "Info",
      VolumeName: "Volume Name",
      Totalcapacity: "Total Capacity",
      Used: "Used",
      UsedRate: "Used%",
      InodeCount: "Inode Count",
      Replications: "Replications",
      OwnerID: "Owner ID",
      CreateTime: "Create Time",
      SecretKey: "Secret Key",
      Comments: "Comments",
      VolumeInfo: "Volume Info",
      VolumeInfoTip:
        'The "Bucket" concept used in Amazon S3 is similar to the concept called "Volume" in ChubaoFS (when using its object storage interfaces).',
      EditVolume: "Edit Volume",
      IncreaseVolumeCapacity: "Increase Volume Capacity",
      Currentcapacity: "Current Capacity",
      Extendto: "Extend to",
      PermissionofVolume_Spark: "Permission of Volume_Spark",
      GrantPermission: "Grant Permission",
      GrantPermissionto: "Grant Permission to",
      EditUserAccess: "Edit User Access",
      UploadFile: "Upload File",
      CreateFolder: "Create Folder",
      FileFolderName: "File/Folder Name",
      Size: "Size",
      LastModify: "Last Modify",
      MD5: "MD5",
      FileName: "File Name",
      ExpirationTime: "Input the expiration time (s)",
      SignURL: "Sign URL",
      DragTip: "Drag or drop your files here，or",
      DragTipPlease: "Please follow the rules below when uploading a file",
      Tips1: "1. File name shoud be UTF-8 encoding",
      Tips2: '2. File name cannot begin with "/"',
      Tips3:
        '3. File name contain letters, numbers, Chinese, "/", ".", "-", and "_"',
      Tips4:
        "4. The length of the file name should be less than 100 characters",
      FolderName: "Folder Name",
      FileDetails: "File Details",
      CreateSignURL: "Create Sign URL",
      EnterNameTip: "Please enter Volume name",
      NameRulesTip:
        "The volume name can contain lowercase letters, numbers, and hyphens. It must start and end with lowercase letters or numbers, and be in the 3-63 length range.",
      EnterCapacityTip: "Please enter Total capacity",
      CapacityRules: "Must be a positive integer",
      FolderNameMsg: "Please enter Folder name",
      deleteInfoMsg: "Are you sure you want to  delete the Folder?",
      dpReplicaNumMsg: "Please select Replications",
      notesMsg: "Please enter Notes",
      userNameMsg: "Please enter User Name",
      accessMsg: "Please select Access",
      updateVolumeMsg:
        "don't support new replicaNum[3] larger than old dpReplicaNum[2]",
      copies: "Number of Copies",
      metadata: "Metadata Store",
      depotDay: "Recycle Bin Days",
      capacity: "Capacity",
      resource: "Special Resources",
      token: "Token Mounting",
      forcedrow: "Forced ROW",
      follow: "follower Read",
      sdk: "SDK Write Cache",
      cache: "Distributed Cache"
    },
    tools: {
      Actions: "Actions",
      Search: "Search ",
      Edit: "Edit ",
      Yes: "Yes",
      No: "No",
      Extend: "Extend",
      Permission: "Permission",
      Comments: "Comments",
      Cancel: "Cancel",
      Create: "Create",
      Remove: "Remove",
      Operations: "Operations",
      Update: "Update",
      Warning: "Warning",
      OK: "OK",
      Access: "Access",
      Details: "Details",
      SignURL: "Sign URL",
      Download: "Download",
      Upload: "Upload",
      Replica: "Replica",
      Add: "Add",
      Delete: "Delete",
      Decommission: "Decommission",
      Depot: "Depot",
      recover: "Recover",
      ClearAll: "One Click Clear",
      Health: "Health",
      Normal: "Normal",
      Emergent: "Emergent"
    },
    commonAttr: {
      UserName: "UserName"
    },
    commonTxt: {
      eachPageShows: "Each page shows"
    },
    timeInterval: {
      Daily: "Daily",
      Weekly: "Weekly",
      Monthly: "Monthly"
    },
    message: {
      Success: "Success",
      Warning: "Warning",
      Error: "Error"
    },
    enterRules: {
      checkNumberTip: "Must be a positive integer"
    },
    accessList: {
      "perm:builtin:ReadOnly": "perm:builtin:ReadOnly",
      "perm:builtin:Writable": "perm:builtin:Writable"
    },
    crumb: {
      Server: ["Server"],
      ServerList: ["Server", "Server List"],
      VolumeList: ["Volume List"]
    }
  },
  ...enLocale
};