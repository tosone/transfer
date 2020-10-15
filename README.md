# transfer

下载国外某些文件的时候总会出现，文件很大，下载需要好久的情况，尽管我们有翻墙的服务在。

如果我们有一个可以在服务器上挂起的程序可以将文件静默的下载到我们指定的地方，会极大地方便我们。

### Storage

如果你实现了 uploader 中 Driver 的相关接口，那么同样也支持相应的存储类型，以下是目前支持的存储类型：

- OSS
- COS
- S3
- qiniu
- minio

### API

- 创建下载 POST `/task` 

``` json
{
    "type": "COS", // 下载到的文件将要存储的云存储类型
    "url": "https://tc.tosone.cn/20190703175351.png",
    "randomFilename": true,
    "path": ".",
    "force": true
}
```

- 获取所有任务 GET `/task`

- 获取任务的进度 GET `/task/{name}`

### 设计与实现

##### 数据库

- Task 任务
- URL 短链
- Account 账户
