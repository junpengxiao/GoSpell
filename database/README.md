#Introduction
**database** wraps GAE datastore to support blog engine. It maintains data's link structure. In short, if data have this structure: A1->A2->A3... then you can just put them into database. Querying A1 will return A1 and the key of A2. Since GAE datastore has Quota, database will build a cache to reduce query times.

For example, a basic blog will have 2 types of pages : pages to display article lists (Contents) and pages to display a full article. You can just save every article into one database table, but it is expensive when you build the article lists page because the query returns all articles. Instead you should save title & abstract as A1, and use whole passage as A2, then put [A1 A2] into database. To construct article lists, just query A1, and if you need the full version of article, just use the key (returned with A1) to get A2.

#Interface
##Struct DataStruct
```go
type DataStruct struct {
     Content, NextKey string
     Date time.Time
}
```
Content is the data you want to save. NextKey will be explained below.
##PutData
```go
PutData(content []string, kind string, context appengine.Context) error
```
This function will accpet a slice of string and a kind name specified by user (kind name is used to fetch data). **Example: ** If you put [Str0 Str1 Str2] into database under the name of *article* then the database will put them in 3 tables under the kind name article0, article1, article2 seperately.

|article0|        |article1|        |article2|
|:------:|:------:|:------:|:------:|:------:|
|Str0|-->|Str1|-->|Str2|

**Notice** content number should be greater than 0 and less than or equal with 5

##Get Certain Data
```go
GetData(key string, context appengine.Context) (DataStruct, error)
```
This function will return the SataStruct specified by key

##Query Data
```go
QueryData(kind string, level, skip, number int, context appengine.Context) ([]DataStruct, error)
```

This function will execute a query and return a slice of DataStruct. In the **Example above** the kind should be *article* and the level should be 0 or 1 or 2, corresponding to the first table, second table, third table. To illustrate other paras, let's draw a new table.

|article0|
|:------:|
|abc|
|def|
|ghi|

`skip` tells the query skip a certain number of items before collect the value. `number` tells the query return a number of items. To get 10 itmes start from "def", use `QueryData("article", 0, 1, 10, context)`. But this will only return 2 items due to the table only contains 3 itmes.

##Update Data
```go
func UpdateData(content []string, strkey string, context appengine.Context) (string, error)
```
This function will put content[0] into strkey.Content, content[1] into strkey.NextKey.Content. If nextkey is empty, then function will insert a new item into database, and link nextkey with this item. And will return a key which is equal with strkey.

**Notice** number of content should be greater than 0 and less than or equal with 5. **AND** Yes, the length of link can be greater than 5, if you use UpdateData repeatedly on the link's last node. The "5" limitation is duo to the Google DataStore Cross-Group Transaction Limitation.
##StoreData
```go
func StoreData(data DataStruct, kind string, context appengine.Context) (string ,error)
```
This function will store a data	directly under the kind.
#To Do
- [ ] Add Cache
