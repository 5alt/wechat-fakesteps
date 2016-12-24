# 利用蓝牙模拟手环伪造微信运动步数

伪造微信运动步数的方法有很多种，可以hook系统提供步数统计的模块，可以伪造GPS，可以找第三方手环的漏洞。
第三方手环走微信iot平台[AirSync协议](http://iot.weixin.qq.com/wiki/new/index.html?page=4-2-1)的话可能通信过程启用AES加密，很难对其攻击。于是考虑自己成为一个手环设备生产商并伪造一个运动手环，欺骗微信运动来获取假的运动步数。

## 微信平台配置
此部分内容主要参考[微信iot平台文档](http://iot.weixin.qq.com/wiki/new/index.html?page=3-4-1)。

### 公众号后台开通“设备功能”插件
如果没有认证过的公共号，也可以用公众号测试账号。
http://mp.weixin.qq.com/debug/cgi-bin/sandbox?t=sandbox/login
添加设备时，需要指明接入方案，选平台基础接入方案，连接类型选蓝牙，产品配置选蓝牙发现。登记成功后会得到一个微信硬件的型号编码即product_id。

### 获取access_token
https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid={{appid}}&secret={{secret}}
获取完access_token之后，接下来从微信公共号后台找到自己的openid。

### 获取deviceid和二维码
https://api.weixin.qq.com/device/getqrcode?access_token={{access_token}}&product_id={{product_id}}
得到如下返回。
```
{"base_resp":{"errcode":0,"errmsg":"ok"},"deviceid":"gh_eee2c24a6f8e_852ddc34836c0559","qrticket":"http:\/\/we.qq.com\/d\/AQC5K_3BgSGNsg84sHISxXmHwMJrSp5sDf9AX1sB"}
```
微信平台会分配一个deviceid和对应的二维码qrticket，在绑定用户的时候用到。

### 设备授权
https://api.weixin.qq.com/device/authorize_device?access_token={{access_token}}
POST数据
```
{
    "device_num":"1",
    "device_list":[
    {
        "id":"gh_eee2c24a6f8e_852ddc34836c0559",
        "mac":"ff8d22e19590",
        "connect_protocol":"3",
        "auth_key":"",
        "close_strategy":"1",
        "conn_strategy":"5",
        "crypt_method":"0",
        "auth_ver":"0",
        "manu_mac_pos":"-1",
        "ser_mac_pos":"-2",
        "ble_simple_protocol": "1"
    }
    ],
    "op_type":"0",
    "product_id": "25806"
}
```
主要为开启蓝牙精简协议。在数据包里配置好id、mac和product_id即可。mac为电脑蓝牙的mac地址。协议的具体内容参照[微信文档](http://iot.weixin.qq.com/wiki/new/index.html?page=3-4-5)。


### 强制绑定用户和设备
https://api.weixin.qq.com/device/compel_bind?access_token={{access_token}}
POST数据
```
{
    "device_id": "gh_eee2c24a6f8e_852ddc34836c0559",
    "openid": "ouSvtwWqDVHQUGT3u0XkTpiJ9QsY"
}
```
在数据包里配置好device_id和openid。

这样，微信账号就与设备绑定好了。打开公共号会看到公共号下面有个“未连接”的字符串。

## 伪造手环设备

根据[微信蓝牙精简协议文档](http://iot.weixin.qq.com/wiki/new/index.html?page=4-3)

> 设备需要广播包带上微信的service，并在manufature data里带上mac地址。
> 微信Service uuid：0xFEE7
> manufature specific data：需以MAC地址（6字节）结尾。并且manufature specific data长度需大于等于8字节（最前两个字节为company id，没有的话随便填）。
> 微信service下面需包含一个读特征值，uuid为：0xFEC9，内容为6字节MAC地址（ios系统其他软件连上设备之后，微信会去读该特征值，以确定设备MAC地址）。

所以只需要建立一个uuid为0xFEE7的service，里面包含0xFEA1、0xFEA2、0xFEC9三个Characteristic并设置好值和属性即可。

代码见`server.go`，其中步数和蓝牙的mac地址可以配置。

使用方法：
```
go build server.go
sudo ./server
```
需要高于某个版本的golang。

另外看到国内有人用nodejs实现了[类似的功能](https://github.com/luluxie/weixin-iot)。

注意node版本需要[大于0.12](https://github.com/nodesource/distributions#installation-instructions)，另外bleno版本需要为0.4.0。

```
sudo apt-get install libusb-1.0-0-dev
npm install bleno@0.4.0
```

## 参考文献
http://iot.weixin.qq.com/wiki/new/index.html?page=4-3
https://github.com/luluxie/weixin-iot
https://github.com/paypal/gatt
