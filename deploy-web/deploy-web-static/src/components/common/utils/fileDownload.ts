import axios from "axios";
type Method = "get" | "GET" | "delete" | "DELETE" | "head" | "HEAD" | "options" | "OPTIONS" | "post" | "POST" | "put" | "PUT" | "patch" | "PATCH" | "purge" | "PURGE" | "link" | "LINK" | "unlink" | "UNLINK"


/**
 * @interface
 * @description 获取文件数据配置参数接口定义
 * @param {
    * url?: string 下载路径url
    * method?: Method 接口method
    * fileName?: string 文件名
    * blobType?: string 文件类型
    * params?: any getPath函数传入的参数
    * downloadParams?: any 下载接口额外参数
    * } config 配置项
    */
   interface IFfetchFileData {
       (config: { url: string; method?: Method; blobType?: string; params?: any }): void
   }

/**
* @description: 从指定url获取文件数据并生成文件
* @param {IFfetchFileData} {
*  url: 获取文件路径
*  method: 接口method
*  blobType: 文件类型
* }
*/
export const fetchFileData: IFfetchFileData = async ({
    url,
    method = 'get' as const,
    blobType = 'application/vnd.ms-excel;charset=utf8',
}) => {
    axios.create({
        headers: {
            'Content-Type': 'application/json;charset=UTF-8',
            'Cache-Control': 'no-cache',
            Pragma: 'no-cache',
        },
        responseType: 'blob'
    }).get(url).then((res => {
        const fileName = res.headers['content-disposition'].split('attachment; filename=')[1]
        // 生成文件
        generateFile({
            fileName: decodeURI(fileName),
            content: res.data,
            blobType,
        })
    }))
}

/**
* @function
* @description: 根据数据生成文件
* @param {{
*  content: 文件数据
*  fileName: 生成文件名称
*  blobType: 文件类型
* }}
*/
export const generateFile = ({ fileName, content, blobType }: {
    fileName: string
    content: Blob
    blobType: string
}) => {
    const _fileName = fileName
    const blob = new Blob([content], {
        type: blobType,
    })
    if (navigator.appVersion.toString().indexOf('.NET') > 0) {
        const nav = window.navigator as any
        nav.msSaveOrOpenBlob(blob, _fileName)
        return
    }
    const link = document.createElement('a')
    link.innerHTML = _fileName
    link.download = _fileName
    link.href = URL.createObjectURL(blob)
    document.body.appendChild(link)
    const evt = document.createEvent('MouseEvents')
    evt.initEvent('click', false, false)
    link.dispatchEvent(evt)
    document.body.removeChild(link)
}
