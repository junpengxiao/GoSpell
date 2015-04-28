//This database is designed to storing my blog data.
//There are 3 main functions:
//PutData, QueryData, GetData will operate data as name suggestted
//Scheme: The the data should be translate into a string
//and will be stored in DataStruct Content property.
//The data should be given a kind,
//and QueryData will list a subset of datas in a name specified by caller.
//Additionally, You can put a slice of data. All data will be linked via RefKey.
//i.e., program will first store data[end]
//then assign data[end]'s key to data[end-1]'s refkey property.
//[data0,data1,data2] <==> [data0]-->[data1]-->[data2]
//This is a situation that may use the scheme
//If we build a blog, we need 2 pages:
//one lists all article title and abstract, the other shows full article. 
//When user click the title, we can display the whole article
//In this case, we will put [title & abstract, title & content] into datastore.
//Data[0], i.e., title & abstract can be queried.
//if we need data[1], we call GetData and pass it data[0].RefKey
//Ofcourse you can also query data[1] directly
package database

import (
	"appengine"
	"appengine/datastore"
	"time"
	"strconv"
	"log"
)

type DataStruct struct {
	Content, NextKey string `datastore:",noindex"`
	Date time.Time
}

func kindRootKey(kind string, context appengine.Context) *datastore.Key {
	return datastore.NewKey(context, kind+"Root", kind+"Root", 0, nil)
}

func StoreData(data *DataStruct, kind string, context appengine.Context) (string ,error) {
	if key, err := datastore.Put(context,
		datastore.NewIncompleteKey(context, kind,
			kindRootKey(kind, context)),
		data); err != nil {
			log.Println("Error in database, storeData, ", err)
			return "", err
		} else {
			return key.Encode(), nil
		}
}

//This function will put []string into datastore, data[i] will be stored in "kindi" field, i.e., kind + i.toString.
func PutData(content []string, kind string, context appengine.Context) (string, error) {
	if len(content) == 0 || len(content) > 5 {
		log.Println("Warning In database, PutData, content number is 0 or larger than 5")
		return "",nil
	}
	if kind == "" || '0'<=kind[len(kind)-1] && kind[len(kind)-1]<='9' {
		log.Println("Warning In database, Putdata, kind last char cannot be digit")
		return "",nil
	}
	transOpt := datastore.TransactionOptions { XG : true,}//cross-group
	data := make([]DataStruct, len(content))
	for i := 0; i != len(content); i++ {
		data[i].Content = content[i]
		data[i].Date = time.Now()
		data[i].NextKey = ""
	}
	var key0 string;
	err := datastore.RunInTransaction(context,
		func(context appengine.Context) error {
			var err2 error
			for i := len(data)-1; i!=0; i-- {
				data[i-1].NextKey,err2 = StoreData(&data[i],kind + strconv.Itoa(i),context)
				if err2 != nil {
					return err2
				}
			}
			key0, err2 = StoreData(&data[0],kind + strconv.Itoa(0),context);
			if len(data)>= 2 {
				data[len(data)-1].NextKey = key0
				dataKey, _ := datastore.DecodeKey(data[len(data)-2].NextKey)
				datastore.Put(context, dataKey, &data[len(data)-1])
			} else {
				data[0].NextKey = key0
				dataKey, _ := datastore.DecodeKey(key0)
				datastore.Put(context, dataKey, &data[0])
			}
			return err2
		}, &transOpt)
	if err == datastore.ErrConcurrentTransaction {
		log.Println("Error in database, PutData, transaction failed")
	} else if err != nil {
		log.Println("Warning in database, PutData, transction warning", err)
	}
	return key0, err
}

//GetData will accept a key and return the content pointed by that key
func GetData(key string, context appengine.Context) (DataStruct, error) {
	var rect DataStruct
	dataKey, err := datastore.DecodeKey(key)
	if (err != nil) {
		log.Println("Error in database, GetData, decodekey", err)
		return rect, err
	}
	err = datastore.Get(context, dataKey, &rect);
	if err != nil {
		log.Println("Error in database, GetData, get content", err)
	}
	return rect, err
}

//QueryData will return a list of content. 'kind' is as same as kind in PutData.
//level is the column you query. recall that put data will put
//[d00, d01, d02] 
//[d10, d11, d12]
//datastore via a link form.
//kind0  kind1  kind2
//[d00]->[d01]->[d02]
//[d10]->[d11]->[d12]
//if you want to get the first column, i.e., d00,d10,d20... then level should be 0
//skip is the number of item  you will skip. if you set skip = 1, then result will begin with d10
//number means how many item do you want to get. 
func QueryData(kind string, level, skip, number  int, context appengine.Context) ([]DataStruct, error) {
	if number <=0 {
		log.Println("Warning in database, QueryData, number < 0")
		return nil, nil
	}
	if skip<0 {
		skip = 0
	}
	query := datastore.NewQuery(kind + strconv.Itoa(level)).
		Ancestor(kindRootKey(kind + strconv.Itoa(level), context)).
		Offset(skip).Limit(number).Order("-Date")
	result := make([]DataStruct, 0, number)
	_ , err := query.GetAll(context, &result);
	if err != nil {
		log.Println("Error in database, QueryData, ", err)
		return nil, nil
	}
	
	return result,  nil
}

//UpdateData will store a list of contents into a list begin with strkey.
//It will 1st fetch an item by the key of strkey and then revive the list
//begin with that item.
//strkey->strkey.NextKey->...
//until nextkey is empty or the length of list is equal with len(content)
//Finally the function will put the list into datastore's same position
func UpdateData(content []string, strkey string, context appengine.Context) (string, error) {
	key, err := datastore.DecodeKey(strkey)
	if err != nil {
		log.Println("Error in database, updatedata, ", err)
		return "", err
	}
	kind := key.Kind()
	level := 0
	for base := 1; '0'<=kind[len(kind)-1] && kind[len(kind)-1]<='9'; kind = kind[0:len(kind)-1] {
		level += base * int(kind[len(kind)-1]-'0')
		base *= 10
	}
	if len(content) == 0 || len(content) > 5 {
		log.Println("Warning in database, UpdateData, content number is 0 or larger than 5")
		return "", nil
	}
	transOpt := datastore.TransactionOptions {XG : true,} //cross - group
	data := make([]DataStruct, len(content))
	for i, nowkey :=0, strkey; i != len(content); i++ {
		data[i], _ = GetData(nowkey, context)
		if data[i].NextKey == "" {
			break
		} else {
			nowkey = data[i].NextKey
		}
	}
	for i := 0; i != len(content); i++ {
		data[i].Content = content[i]
		data[i].Date = time.Now()
	}
	level += len(data)
	var key0 string
	err = datastore.RunInTransaction(context,
		func(context appengine.Context) error {
			var err2 error
			for i := len(data)-1; i!=0; i--{
				level -= 1
				if data[i-1].NextKey != "" {
					key, _ := datastore.DecodeKey(data[i-1].NextKey)
					key, err2 = datastore.Put(context, key, &data[i])
					if (err2 != nil) {
						log.Println("Error in database, updatedata, update original content ", err2)
						return err2
					}
					data[i-1].NextKey = key.Encode()
				} else {
					data[i-1].NextKey, err2 = StoreData(&data[i], kind + strconv.Itoa(level), context)
					if err2 != nil {
						return err2
					}
				}
			}
			log.Println("data ", data)
			if tmpkey, err := datastore.Put(context, key, &data[0]); err != nil {
				//key is decoded from original para.
				log.Println("Error in database, updatedata, 1st content err ", err)
				return err
			} else {
				key0 = tmpkey.Encode()
			}
			return nil
		}, &transOpt)
	if err == datastore.ErrConcurrentTransaction {
		log.Println("Error in database, Updatedata, transaction failed")
	} else if err != nil {
		log.Println("Warning in database, Updatedata, transaction warning", err)
	}
	return key0, err
}
