# 数据库设计

此次的数据库分为三个表：

* 用户信息
* 委托信息
* 问卷信息

## 用户信息

用户信息的表主要包括：

|字段|类型|解释|
|--|--|--|
|_id|string|对象的id|
|open_id|string|用于其他表格中的用户id|
|name|string|用户名|
|student_num|string|学号|
|credit|int|用户的积分，只能为正|

## 委托信息

委托信息的表主要包括：

|字段|类型|解释|
|--|--|--|
|_id|string|对象的id|
|publisher_id|string|发布者的id|
|receiver_id|string|接受者的id|
|delegation_name|string|委托的名字|
|start_time|int64|开始的时间，Unix时间戳|
|delegation_state|4|委托的状态|
|reward|5|委托的积分奖励|
|description|string|委托的描述|
|deadline|int64|委托结束的时间，Unix时间戳|
|delegation_type|string|委托的类型|

还包括一些只有包含问卷的委托才会用上的字段：

|字段|类型|解释|
|--|--|--|
|questionnaire_id|string|对应的问卷的id|
|max_number|int|问卷的最高填写人数|
|current_number|int|问卷的当前填写人数|

## 问卷信息

问卷信息的表主要包括：

### 问卷表

|字段|类型|解释|
|--|--|--|
|_id|string|对象的id|
|questions|array|问题数组|

### 问题的组成

```
- questions         -问题，数组
    |
    -topic          -问题的标题
    -answers        -选项和统计，数组
        |
        -option     -选项
        -number     -选择此选项的人数统计
```
