package database
import (
	"testing"
	"appengine/aetest"
	"appengine/datastore"
	"time"
)

//This function will test StoreData via a single item
//content: 134 kind: test Date: date.now().round()
func TestStoreData(t *testing.T) {
	context, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer context.Close()
	content := "134"
	date := time.Now().Round(time.Second)
	data := DataStruct{
		content,
		"",
		date,
	}
	kind := "test"
	var key0 string
	if key0, err = StoreData(&data, kind, context); err != nil {
		t.Fatal(err)
	}
	query := datastore.NewQuery(kind).Ancestor(kindRootKey(kind,context))
	result := make([]DataStruct, 0, 10)
	if _, err := query.GetAll(context, &result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || data != result[0] {
		t.Errorf("BadQueryResult ",result )
	}
	datakey, err2 := datastore.DecodeKey(key0)
	if err2 != nil {
		t.Errorf("store data returns wrong key")
	}
	err = datastore.Get(context, datakey, &result[0])
	if (err != nil) {
		t.Fatal(err)
	}
	if (result[0] != data) {
		t.Errorf("Query by key error", result[0])
	}
}

func TestPutDataStructure(t *testing.T) {
	context, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer context.Close()
	content := []string{"123", "456", "789"}
	key, err := PutData(content, "test", context)
	if err != nil {
		t.Errorf("Error while put data ", err)
	}
	data, _:= GetData(key, context)
	if data.Content != content[0] {
		t.Errorf("dataStructure 1st data wrong", data.Content)
	}
	data, _ = GetData(data.NextKey, context)
	if data.Content != content[1] {
		t.Errorf("dataStructure 2nd data wrong", data.Content)
	}
	data, _ = GetData(data.NextKey, context)
	if data.Content != content[2] {
		t.Errorf("dataStructure 3rd data wrong", data.Content)
	}
	data, _ = GetData(data.NextKey, context)
	if data.Content != content[0] {
		t.Errorf("data circle structure is wrong", data.Content)
	}
}

//This function first store ["123", "456", "789"] 5 times under 'test' kind. And then
//Query test0, result should be "123" 5 times. And then check NextKey
//Finally Query test1, result should be "456" 5times
func TestPutData(t *testing.T) {
	const repeat = 5;
	context, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer context.Close()
	content := []string{"123", "456", "789"}
	var key0 string
	for i:= 0; i!= repeat; i++ {
		key0, err = PutData(content, "test", context)
		if err != nil {
			t.Fatal(err)
		}
	}
	query := datastore.NewQuery("test0").Ancestor(kindRootKey("test0", context))
	result := make([]DataStruct, 0, 10)
	datakey, tmperr := datastore.DecodeKey(key0)
	if tmperr != nil {
		t.Fatal(err)
	}
	var tmpresult DataStruct
	err = datastore.Get(context, datakey, &tmpresult)
	if tmpresult.Content != content[0] {
		t.Errorf("Result query by key0 error")
	}
	if _, err := query.GetAll(context, &result); err != nil {
		t.Fatal(err)
	}
	if len(result)!=repeat {
		t.Errorf("In TestPutData, Result len is not right ", result)
	}
	for _, data := range result {
		for cnt:= 0; cnt !=  len(content);cnt++ {
			if data.Content != content[cnt] {
				t.Errorf("In TestPutData, Bad Query Result ", data)
			}
			key, tmperr := datastore.DecodeKey(data.NextKey)
			if tmperr != nil {
				t.Fatal(tmperr)
			}
			if err := datastore.Get(context, key, &data); err !=nil {
				t.Fatal(err)
			}
		}
	}
	query = datastore.NewQuery("test1").Ancestor(kindRootKey("test1", context))
	result2 := make([]DataStruct, 0, 10)
	if _, err := query.GetAll(context, &result2); err != nil {
		t.Fatal(err)
	}
	if len(result2) != repeat {
		t.Error("In TestPutData, test1 kind result length is not right")
	}
	for _, data := range result2 {
		if data.Content != content[1] {
			t.Errorf("In TestPutData, Bad Result in test1 kind", result2)
		}
	}
}

//This test will first store "123" under test0, and "456" under test1. Then use key in
//"123" to fetch "456"
func TestGetData(t *testing.T) {
	context, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer context.Close()
	content := []string{"123", "456"}
	firstkey, err := PutData(content, "test", context)
	if err != nil {
		t.Fatal(err)
	}
	query := datastore.NewQuery("test0").Ancestor(kindRootKey("test0", context))
	result := make([]DataStruct, 0, 10)
	if _, err := query.GetAll(context, &result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Errorf("result length is not 1")
	}
	if result1, err := GetData(result[0].NextKey, context); err != nil {
		t.Fatal(err)
	} else if result1.Content != content[1] {
		t.Errorf("Result is not correct", result1)
	} else if result1.NextKey != firstkey {
		t.Errorf("Result1 next key should be empty string")
	}
}

//Put ["123","456","789"] 5 times. check skip, num limit, multi-kind-level, and result
func TestQueryData(t *testing.T) {
	context, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer context.Close()
	content := []string{"123", "456", "789"}
	const repeat = 5
	for i := 0; i != repeat; i++ {
		if _, err := PutData(content, "test", context); err != nil {
			t.Fatal(err)
		}
	}
	result, _ := QueryData("test", 0, 1, 5, context)
	if len(result) != 4 {
		t.Errorf("In TestQueryData, skip error. ", result)
	}
	for _, data := range result {
		if data.Content != content[0] {
			t.Errorf("In TestQueryData, test 0 content error")
		}
		doubleCheck, _ := GetData(data.NextKey, context)
		if doubleCheck.Content != content[1] {
			t.Errorf("In TestQueryData, test 0 NextKey error")
		}
	}
	result, _ = QueryData("test", 2, 0, 3, context)
	if len(result) != 3 {
		t.Errorf("In TestQueryData, number limit error", result)
	}
	for _, data := range result {
		if data.Content != content[2] {
			t.Errorf("QueryData, test 2 content error")
		}
	}
}

//put ['123','456', '789'] under the kind of "test",
//update '456' with ['abc', 'edf'], thus data should be
//test0   test1   test2  
//'123'   'abc'   'def'  
//retrieve the link by query test0
//check test1
//check test0
func TestUpdateData(t *testing.T) {
	context, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer context.Close()

	content := []string{"123","456","789"}
	strkey, DebugErr := PutData(content,"test",context)
	if DebugErr != nil {
		t.Errorf("DebugErr: ", DebugErr)
	}
	newcontent := []string{"abc","def"}
	data, _ := GetData(strkey, context)
	checkKey, _ := UpdateData(newcontent,data.NextKey, context)
	if checkKey != data.NextKey {
		t.Errorf("UpdateData, Assumption Wrong")
	}
	//retrieve full link
	for i:= 0; i != 2; i++ {
		data, _ = GetData(data.NextKey, context)
		if data.Content != newcontent[i] {
			t.Errorf("UpdateData, new content mis-match ", data)
		}
	}
	//check test1
	result, _ := QueryData("test", 1, 0, 1, context)
	if result[0].Content != newcontent[0] {
		t.Errorf("UpdateData, kind 1 mis-match ", result)
	}
	//check test0
	result, _ = QueryData("test", 0, 0, 1, context)
	if result[0].Content != content[0] {
		t.Errorf("UpdateData, kind 3 mis-match ", result)
	}

}
