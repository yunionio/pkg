通过struct field的指针地址获取struct的fieldname
================================================

使用方法：

为每一个schema申明一个全局的静态变量，这个变量只用来干找到fieldname这个事情。

例如：

    var Network SNetwork
    func init() {
        sqlchemy.RegisterStructFieldNames(&Network)
    }

后面则用fieldname.Fn(&Network.GuestDhcp)获得GuestDhcp这个字段对应的column name
