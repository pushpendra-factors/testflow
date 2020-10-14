package operations

import(
	"testing"
	"reflect"
)

func TestConvert1(t *testing.T){
	ip := make(map[string]interface{})
	ip["a"] = 0.1
	ip["b"] = 0.2
	result := Convert1(ip)
	if(result["a"] != 0.1){
		t.Errorf("Expected: 0.1 Result: %v", result["a"])
	}
	if(reflect.TypeOf(result).Kind() != reflect.Map){
		t.Errorf("Expected: %v Result: %v", 
			reflect.Map, 
			reflect.TypeOf(result).Kind())
	}
	if(reflect.TypeOf(result["a"]).Kind() != reflect.Float64){
		t.Errorf("Expected: %v Result: %v", 
			reflect.Float64, 
			reflect.TypeOf(result["a"]).Kind())
	}
}

func TestConvert2(t *testing.T){
	ip := make(map[interface{}]interface{})
	ip["a"] = 0.1
	ip["b"] = 0.2
	result := Convert2(ip)
	if(result["a"] != 0.1){
		t.Errorf("Expected: 0.1 Result: %v", result["a"])
	}
	for item, element := range result {
		if(reflect.TypeOf(item).Kind() != reflect.String){
			t.Errorf("Expected %v, Result %v", 
				reflect.String,
				reflect.TypeOf(item).Kind())
		}
		if(reflect.TypeOf(element).Kind() != reflect.Float64){
			t.Errorf("Expected %v, Result %v", 
				reflect.Float64,
				reflect.TypeOf(element).Kind())
		}
		break
	}
}