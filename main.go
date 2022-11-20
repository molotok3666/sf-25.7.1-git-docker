package main

import (
	"fmt"
	"sync"
	"time"
)

const BUFFER_SIZE = 5             // Размер буффера
const BUFFER_CLEAR_INTERFVAL = 10 // Интервал очистки буффера

type ringBuffer struct {
	arr  []int
	pos  int
	size int
	m    sync.Mutex
}

// Добавляет элемент в буффер
func (r *ringBuffer) Push(el int) {
	r.m.Lock()
	defer r.m.Unlock()

	if r.pos == r.size-1 {
		for i := 1; i <= r.size-1; i++ {
			r.arr[i-1] = r.arr[i]
		}
	} else {
		r.pos++
	}

	r.arr[r.pos] = el
}

// Берет элемент из буффера (FIFO)
func (r *ringBuffer) Get() []int {
	if r.pos < 0 {
		return nil
	}

	r.m.Lock()
	defer r.m.Unlock()
	output := r.arr[:r.pos+1]
	r.pos = -1

	return output
}

// Создает новый буффер
func NewRingBuffer() *ringBuffer {
	return &ringBuffer{make([]int, BUFFER_SIZE), -1, BUFFER_SIZE, sync.Mutex{}}
}

// Фильтрует отрицательные числа
func filterNegative(done <-chan bool, ch <-chan int) <-chan int {
	nextCh := make(chan int)

	go func() {
		for {
			select {
			case number := <-ch:
				if number >= 0 {
					nextCh <- number
				}
			case <-done:
				return
			}
		}
	}()

	return nextCh
}

// Фильтрует числа не кратные 3 и равные 0
func filterNotDivThree(done <-chan bool, ch <-chan int) <-chan int {
	nextCh := make(chan int)
	go func() {
		for {
			select {
			case number := <-ch:
				if number != 0 && number%3 == 0 {
					nextCh <- number
				}
			case <-done:
				return
			}
		}
	}()

	return nextCh
}

// Записывает в буффер
func writeToBuffer(done <-chan bool, ch <-chan int) <-chan int {
	nextCh := make(chan int)
	buffer := ringBuffer{make([]int, BUFFER_SIZE), -1, BUFFER_SIZE, sync.Mutex{}}

	go func() {
		for {
			select {
			case number := <-ch:
				buffer.Push(number)
			case <-done:
				return
			}
		}
	}()

	go func() {
		for {
			time.Sleep(BUFFER_CLEAR_INTERFVAL * time.Second)
			bufferData := buffer.Get()

			for _, data := range bufferData {
				select {
				case nextCh <- data:
				case <-done:
					return
				}
			}
		}
	}()

	return nextCh
}

// Возвращает источник данных (канал чтения из консоли)
func getSource() (<-chan int, <-chan bool) {
	ch := make(chan int)
	done := make(chan bool)

	go func() {
		defer close(done)
		var number int

		for {
			_, err := fmt.Scanf("%d\n", &number)
			if err != nil {
				fmt.Println("Please anter an integer number")
				continue
			}

			ch <- number
		}

	}()

	fmt.Println("Введите числа")

	return ch, done
}

func main() {
	source, done := getSource()
	bufferedPipeline := writeToBuffer(done, filterNotDivThree(done, filterNegative(done, source)))

	for data := range bufferedPipeline {
		fmt.Printf("Получены данные: %d\n", data)
	}
}
