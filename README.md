# transfer

下载国外某些文件的时候总会出现，文件很大，下载需要好久的情况，尽管我们有翻墙的服务在。

如果我们有一个可以在服务器上挂起的程序可以将文件静默的下载到我们指定的地方，会极大地方便我们。

## Cloud Storage

如果你实现了 uploader 中 Driver 的相关接口，那么同样也支持相应的存储类型，以下是目前支持的存储类型：

- [OSS](https://cn.aliyun.com/product/oss)
- [COS](https://cloud.tencent.com/product/cos)
- [S3](https://aws.amazon.com/cn/s3/)
- [qiniu](https://www.qiniu.com/products/kodo)
- [minio](https://min.io/)

### API

- 创建下载 POST `/tasks`

``` javascript
{
    "downloadType": "simple", // 下载到的文件将要存储的云存储类型
    "downloadUrl": "https://tc.tosone.cn/20190703175351.png", // 下载文件的 URL
    "randomFilename": true, // 是否采用随机的文件名
    "path": ".", // 下载后在云存储上的路径
    "force": true, // 发现有相同的 URL 后是否强制重新下载
    "uploadType": "qiniu" // 上传类型
}
```

- 获取所有任务 GET `/tasks`

``` javascript
[
    {
        "name": "2a7d8e78c98fbe8c",
        "uploadType": "qiniu",
        "url": "https://tc.tosone.cn/20190703175351.png",
        "downloadType": "simple",
        "downloadUrl": "https://tc.tosone.cn/test/test.png",
        "filename": "test.png",
        "randomFilename": false,
        "path": "./test",
        "force": true,
        "progress": 100,
        "status": "done",
        "message": "",
        "createdAt": "2020-12-29T16:40:23.802822+08:00",
        "updatedAt": "2020-12-29T16:40:25.929417+08:00"
    }
]
```

- 获取任务的进度 GET `/tasks/{name}`

``` javascript
{
    "name": "2a7d8e78c98fbe8c",
    "uploadType": "qiniu",
    "url": "https://tc.tosone.cn/20190703175351.png",
    "downloadType": "simple",
    "downloadUrl": "https://tc.tosone.cn/test/test.png",
    "filename": "test.png",
    "randomFilename": false,
    "path": "./test",
    "force": true,
    "progress": 100,
    "status": "done",
    "message": "",
    "createdAt": "2020-12-29T16:40:23.802822+08:00",
    "updatedAt": "2020-12-29T16:40:25.929417+08:00"
}
```
