package main

import (
	"fmt"
	lrukreplacer "lruKReplacer/lru_k_replacer"
)

func main() {
	var cacheCapacity int = 3
	var timestampsRegisterCapacity = 2
	var cache = lrukreplacer.InitCache(cacheCapacity, timestampsRegisterCapacity)

	var result int
	var err error

	if result, err = cache.Set(1, 10); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)

	if result, err = cache.Set(2, 20); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)

	if result, err = cache.Set(3, 30); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)

	if result, err = cache.Get(1); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)

	if result, err = cache.Get(2); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)

	if result, err = cache.Set(2, 200); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)

	if result, err = cache.Get(3); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)

	if result, err = cache.Set(3, 300); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)

	if result, err = cache.Get(2); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)

	if result, err = cache.Get(3); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)

	if result, err = cache.Get(1); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)

	if result, err = cache.Set(4, 40); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)

	if result, err = cache.Get(4); err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Println("result: ", result)
}
