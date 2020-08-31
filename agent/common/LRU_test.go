package common

import (
	"fmt"
	"testing"
)

func TestLRUCache(t *testing.T) {
	lru := NewLRUCache(3)
	//lru.Set(10,"value1")
	//lru.Set(20,"value2")
	//lru.Set(30,"value3")
	//lru.Set(10,"value4")
	//lru.Set(50,"value5")
	//// 5,1,3 20是没有了

	arr1 := []string{"123","3244","2323"}
	arr2 := []string{"xx","xx","xx"}
	lru.Set(10,arr1)
	lru.Set(20,arr2)
	lru.Set(30,"value3")


	fmt.Println("lru size:",lru.Size())
	v,ret,_:= lru.Get(10)
	if ret {
		fmt.Printf("get true:%v %T\n",v,v)
		for _,i := range v.([]string){
			fmt.Printf("%s ",i)
		}
		fmt.Printf("\n")
	}else{
		fmt.Println("not exit")
	}

	if lru.Remove(30){
		fmt.Println("true")
	}else {
		fmt.Println("false")
	}
	fmt.Println("lru size:",lru.Size())
}
