package main

import "log"

type Base1 struct {
}

func (b *Base1) Foo() {
	log.Printf("Base1.Foo")
}

type Base2 struct {
}

func (b *Base2) Foo() {
	log.Printf("Base1.Foo")
}

type Derived struct {
	Base1
	Base2
}

func (d *Derived) X() {

}

func main() {
	var f Derived
	f.X()
	//f.Foo()
}
