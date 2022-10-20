package main

import (
	"sync/atomic"
)


type Result[A any] struct {
	res        A
	err        error
	errChecked *int32 // use with sync.AtomicInt32 where > 0 is true
}

func Lift[A any](res A, err error) *Result[A] {
	if err != nil {
		return Err[A](err)
	}
	return OK(res)
}

func OK[A any](res A) *Result[A] {
	return &Result[A]{
		res: res,
		errChecked: new(int32),
	}
}

func Err[A any](err error) *Result[A] {
	return &Result[A]{
		err: err,
		errChecked: new(int32),
	}
}

func (r *Result[A]) IsErr() bool {
	atomic.StoreInt32(r.errChecked, 1)
	return r.err != nil
}

func (r *Result[A]) Err() error {
	atomic.StoreInt32(r.errChecked, 1)
	return r.err
}


func (r *Result[A]) OK() A {
	if atomic.LoadInt32(r.errChecked) != 1 {
		panic("must check for error before unboxing value")
	}
	if r.err != nil {
		panic("can't get value of an error result")
	}
	return r.res
}
