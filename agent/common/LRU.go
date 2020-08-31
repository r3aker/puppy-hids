package common

import (
	"container/list"
	"errors"
)

// list中保存的结构体的格式
type CacheNode struct {
	Key,Value interface{}
}
func (cnode *CacheNode) NewCacheNode(k, v interface{}) *CacheNode {
	return &CacheNode{k,v}
}

type LRUCache struct {
	Capacity int
	dlist    *list.List
	CacheMap map[interface{}] *list.Element
}

func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		Capacity: cap,
		dlist:    list.New(),
		CacheMap: make(map[interface{}]*list.Element),
		// list.Element 作用是 保存数据指针（value)
		// CacheMap[key]*list.Element 是映射key -> value 在list中的位置
	}
}

func (lru *LRUCache) Size() int {
	return lru.dlist.Len()
}

func (lru *LRUCache) Set(k, v interface{}) error {
	if lru.dlist == nil {
		// 使用前要先NewLRUCache
		return errors.New("LRUCache not initialize")
	}

	// 已经存在move to front
	if pElement,ok := lru.CacheMap[k];ok {
		lru.dlist.MoveToFront(pElement)
		pElement.Value.(*CacheNode).Value = v // 更新指针指向的cache值
		return nil
	}

	// 新进来new 一个cache node，返回指针插入到 list 里面
	newElement := lru.dlist.PushFront(&CacheNode{k,v})// list 保存的是CacheNode的指针
	lru.CacheMap[k] = newElement

	if lru.dlist.Len() > lru.Capacity {
		lastElement := lru.dlist.Back()
		if lastElement == nil {
			return nil
		}

		cacheNode := lastElement.Value.(*CacheNode)// 类型转换 ｜ reflect反射
		// 增加pid -> [socket inode...]维护的 socket inode -> pid | pid -> info的删除
		deletePID(cacheNode.Key.(int),cacheNode.Value.([]string))
		delete(lru.CacheMap,cacheNode.Key) // 删除缓存表
		lru.dlist.Remove(lastElement)      // list 里面不再保存 element 的指针 | go gc 进行自动垃圾回收
	}
	return nil
}

// ret 代表是否成功获取到value的值
func (lru *LRUCache) Get(k interface{}) (v interface{},ret bool,err error) {
	if lru.CacheMap == nil {
		return v,false,errors.New("LRUCache not initialize")
	}
	if pElement, ok := lru.CacheMap[k];ok{
		// move to front
		lru.dlist.MoveToFront(pElement)
		return pElement.Value.(*CacheNode).Value,true,nil
	}
	return v,false,nil
}

func (lru *LRUCache) Remove(k interface{}) bool {
	if lru.CacheMap == nil {
		return false
	}

	if pElement,ok := lru.CacheMap[k];ok {
		cacheNode := pElement.Value.(*CacheNode)
		// 增加pid -> [socket inode...]维护的 socket inode -> pid | pid -> info的删除
		deletePID(cacheNode.Key.(int),cacheNode.Value.([]string))
		delete(lru.CacheMap,cacheNode.Key)
		lru.dlist.Remove(pElement)
		return true
	}

	return false
}

func deletePID(pid int,sockets []string){
	for _,v := range sockets{
		delete(LocalSocketInodePID,v)
	}
	delete(LocalPIDInfo,pid)
}