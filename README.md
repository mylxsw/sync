# Sync

Sync Definition:

    - name: sync-logs // 同步定义名，必填
      from: 127.0.0.1:8818 // 远端服务器IP:PORT
      files: // 待同步的文件组，必填
      - src: /var/log // 源文件（夹）
        dest: /tmp/logs // 目标文件（夹）
        delete: true // 是否删除文件，默认false，设置为true，则会对比本地和远端，删除本地多余文件
        ignores: // 忽略的文件，gitignore 格式
        - .git/
        - .DS_Store
        before: 
        - action: command // 动作类型
          command: pwd // action=command 命令
          parse_template: true // action=command 解析command作为模板 
          body: "" // action=dingding，钉钉通知内容
          token: "" // action=command 钉钉群 TOKEN
        after:
        - action: command // 动作类型
          command: pwd // action=command 命令
          parse_template: true // action=command 解析command作为模板 
          body: "" // action=dingding，钉钉通知内容
          token: "" // action=command 钉钉群 TOKEN
        error:
        - action: command // 动作类型
          command: pwd // action=command 命令
          parse_template: true // action=command 解析command作为模板 
          body: "" // action=dingding，钉钉通知内容
          token: "" // action=command 钉钉群 TOKEN
        skip_when_error: true // 发生错误时跳过
      before: // 分组同步前置动作
      - action: command // 动作类型
        command: pwd // action=command 命令
        parse_template: true // action=command 解析command作为模板 
        body: "" // action=dingding，钉钉通知内容
        token: "" // action=command 钉钉群 TOKEN
      after: // 分组同步后置动作
      - action: command
        command: curl -i https://www.baidu.com
      error: // 同步错误动作
      - action: dingding
        body: "## Server {{ sysinfo \"hostname\" }} : {{ .FileSyncGroup.Name }} Has errors\n\n**IP:**
          {{ sysinfo \"ip\" }}\n\n**ERR:** \n\n    {{ .Err }}\n"
        token: YOUR_DINGDING_GROUP_TOKEN


Sync Action Template:

| Structure | Type | Desc | 
| --- | --- | --- |
| .JobID | string | Job ID |
| .FileNeedSyncs | FileNeedSyncs | 需要同步的文件列表 |
| .FileSyncGroup | FileSyncGroup | 同步定义 |
| .Units | []SyncUnit | 待同步的文件列表，全量 |
| .Err | error | 同步错误信息 |
| FileNeedSyncs.Files | []FileNeedSync | 需要同步的文件列表 |
| FileNeedSync.SaveFilePath | string | 同步文件保存路径 | 
| FileNeedSync.SyncOwner | bool | 是否需要同步属主 |
| FileNeedSync.SyncFile | bool | 是否需要同步文件 |
| FileNeedSync.Chmod | bool | 是否需要同步文件权限 |
| FileNeedSync.Type | protocol.Type | 文件类型（普通文件，目录，符号链接）|
| FileNeedSync.RemoteFile | *protocol.File | 源文件信息 |
| FileSyncGroup.Name | string | 同步定义名称 |
| FileSyncGroup.Files | []File | 待同步的文件对 |
| FileSyncGroup.From | string | 同步服务器IP:PORT |
| FileSyncGroup.Token | string | 同步 Token |
| FileSyncGroup.Before | []SyncAction | 同步组前置任务 |
| FileSyncGroup.After | []SyncAction | 同步组后置任务 |
| FileSyncGroup.Error | []SyncAction | 同步组失败任务 |
| SyncAction.Action | string | 执行的动作，支持 command/dingding |
| SyncAction.When | string | 同步规则匹配表达式，只有结果为 true 时才执行该动作，默认为 true |
| SyncAction.Command | string | Action=command 时有效，终端命令，默认直接作为文本，使用 `sh -c` 执行 |
| SyncAction.ParseTemplate | bool | Action=command 时有效，为 true 时解析 Command 为模板 |
| SyncAction.Body | string | Action=dingding 时有效，钉钉通知主体，markdown 格式，支持模板 |
| SyncAction.Token | string | Action=dingding 时有效，钉钉群 TOKEN |
| SyncUnit.Files | []*protocol.File | 待同步的文件元信息，全量 |
| SyncUnit.FileToSync | File | 当前文件（夹）同步定义 |
| File.Src | string | 源文件名 |
| File.Dest | string | 目标文件名 |
| File.Ignores | []string | 忽略的文件列表，gitignore 样式 |
| File.Before | []SyncAction | 当前文件（夹）同步前置任务 |
| File.After | []SyncAction | 当前文件（夹）同步后置任务 |
| File.Error | []SyncAction | 当前文件（夹）同步失败任务 |
| File.SkipWhenError | bool | 当指定时，如果当前同步文件（夹）失败，则跳过继续执行下一组文件 |
| (protocol.File).Path | string | 文件原始路径 |
| (protocol.File).Checksum | string | 文件校验和 |
| (protocol.File).Size | int64 | 文件大小 |
| (protocol.File).Type | string | 文件类型 |
| (protocol.File).Symlink | string | 如果文件类型为符号链接，这里为符号链接指向的地址 |
| (protocol.File).Mode | uint32 | 文件 Mode |
| (protocol.File).Uid | uint32 | UID |
| (protocol.File).Gid | uint32 | GID |
| (protocol.File).User | string | 文件属主名 |
| (protocol.File).Group | string | 文件属组名 |
| (protocol.File).Base | string | 文件基础目录路径 |
