# 环境搭建

## 一、安装依赖

### 1.1 安装 nodejs

安装 node18

### 1.2 安装依赖

```bash
$ npm install
```

## 二、开发

### 2.1 启动 mock 服务器

```bash
$ npm run serve
```

### 2.2 运行 webpack

```bash
$ npm run dev
```

tips: 
执行npm run serve是起一个本地服务器。
里面有所有的后端接口，我们可以直接在本地mock所有后端接口。
只需要修改payload.js里面的返回值，然后修改express.js的状态码，然后重启服务器即可
