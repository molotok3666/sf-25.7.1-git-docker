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
	addLog(fmt.Sprintf("add element \"%d\" into buffer", el))
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
	addLog("get elements from buffer")
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
	addLog("create ring buffer")

	return &ringBuffer{make([]int, BUFFER_SIZE), -1, BUFFER_SIZE, sync.Mutex{}}
}

// Фильтрует отрицательные числа
func filterNegative(done <-chan bool, ch <-chan int) <-chan int {
	addLog(fmt.Sprintf("initialize negative filter"))
	nextCh := make(chan int)

	go func() {
		for {
			select {
			case number := <-ch:
				addLog(fmt.Sprintf("filter number \"%d\" by not div three filter", number))
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
	addLog(fmt.Sprintf("initialize not div three filter"))
	nextCh := make(chan int)
	go func() {
		for {
			select {
			case number := <-ch:
				addLog(fmt.Sprintf("filter number \"%d\" by not div three filter", number))
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
	addLog(fmt.Sprintf("initialize buffer pipeline"))
	nextCh := make(chan int)
	buffer := ringBuffer{make([]int, BUFFER_SIZE), -1, BUFFER_SIZE, sync.Mutex{}}

	go func() {
		for {
			select {
			case number := <-ch:
				addLog(fmt.Sprintf("write to buffer number \"%d\"", number))
				buffer.Push(number)
			case <-done:
				return
			}
		}
	}()

	go func() {
		for {
			addLog(fmt.Sprintf("sleep for %d start", BUFFER_CLEAR_INTERFVAL))
			time.Sleep(BUFFER_CLEAR_INTERFVAL * time.Second)
			addLog(fmt.Sprintf("sleep for %d end", BUFFER_CLEAR_INTERFVAL))
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
	addLog("get source channel")
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

func addLog(m string) {
	currentTime := time.Now()
	fmt.Println(fmt.Sprintf("[%s] %s", currentTime.Format("2006-01-02 15:04:05.000000000"), m))
}

func main() {
	addLog("program start")
	source, done := getSource()
	bufferedPipeline := writeToBuffer(done, filterNotDivThree(done, filterNegative(done, source)))

	for data := range bufferedPipeline {
		fmt.Printf("Получены данные: %d\n", data)
	}
}
