package main

import (
    "fmt"
    "time"
    "sync"
    // "context"
)

// 1)
// func main() { 
//     fmt.Println("Starting work..")

//     process("task1")
//     process("task2")
//     process("task3")

//     fmt.Println("All work completed!")
// }

// func process(task string) { //будет по очереди начинать и заканчивать все задания
//     fmt.Println("Started", task)
//     time.Sleep(time.Second)
//     fmt.Println("Finished", task)
// }

//2)
// func main() { //главная функция (без функций ожидания) склонна выполняться после того, как начала выполнять горутины и закончила прежде чем они завершили свою работу
//     fmt.Println("Starting work..")

//     process("task1")
//     process("task2")
//     process("task3")

//     tasks := []string{"Task1", "Task2", "Task3"}

//     for _, task := range tasks { 
//         go func() {
//             process(task)
//         }() //никто не гарантирует порядок выполнения горутин. Планировщик решает сам. Процессы и горутины. Потоки находятся в рамках процесса. Поток - часть выполнения программы, дочерняя к основной программе (процессу)
//     }

//     fmt.Println("All work completed!")
// }

// func process(task string) {
//     fmt.Println("Started", task)
//     time.Sleep(time.Second)
//     fmt.Println("Finished", task)
// }




// 3)
// func main() {
//     fmt.Println("Starting work..")
//     tasks := []string{"Task1", "Task2", "Task3"}
//     var wg sync.WaitGroup  //структура, которая может дождаться выполнения процессов этой группы. Это не воркер
//     wg.Add(len(tasks)) //увеличиваем счетчик на длину массива

//     for _, task := range tasks { 
//         go func() {
//             defer wg.Done() //отложенное выполнение
//             process(task)
//         }()
//     }

//     wg.Wait() //функция ожидания, запускающаяся после того, как начали выполняться горутины. Приложение ждет, пока выполнятся все горутины и только потом запускается
// }

// func process(task string) {
//     fmt.Println("Started", task)
//     time.Sleep(time.Second)
//     fmt.Println("finished", task)
// }



// 4)
// func main() {
//     fmt.Println("Starting work..")

//     tasks := []string{"Task1", "Task2", "Task3"}

//     resultChan := make(chan string) //небуферизированный канал. Его размер 0. Нельзя записать пока никто не прочитает
//     defer close(resultChan)
//     for _, task := range tasks { //первый параметр - индекс
//         go func() {
//             process(task, resultChan)
//         }()
//     }

//     for range tasks {
//         result := <-resultChan 
//         fmt.Println(result)
//     } //неправильный подход (перебор тасков)

//     fmt.Println("All work completed!")
// }

// func process(task string, result chan<- string) {
//     fmt.Println("Started", task)
//     time.Sleep(time.Second)
//     result <- fmt.Sprint("Finished", task) //каждый процесс зависает на записи в result (перед этим секунду ждет)
// }

// 5)
// func main() {
//     fmt.Println("Starting work..")

//     tasks := []string{"Task1", "Task2", "Task3"}

//     resultChan := make(chan string, len(tasks)) //буферизированный

//     var wg sync.WaitGroup
//     wg.Add(len(tasks))

//     for _, task := range tasks { 
//         go func() {
//             defer wg.Done()
//             process(task, resultChan)
//         }()
//     }

//     go func() {
//         defer close(resultChan)
//         wg.Wait()
//     }()

//     for result := range resultChan { //range может читать канал без соответствующего указателя
//         fmt.Println(result)
//     }


//     fmt.Println("All work completed!")
// }

// func process(task string, result chan<- string) {
//     fmt.Println("Started", task)
//     time.Sleep(time.Second)
//     result <- fmt.Sprint("Finished", task)
// }

// 6)
// func main() {
//     fmt.Println("Starting work..")

//     tasks := []string{"Task1", "Task2", "Task3"}

//     resultChan := make(chan string, len(tasks))
//     defer close(resultChan)

//     for i, task := range tasks { 
//         go func() {
//             process(task, resultChan, i+1) //результат выполнения функции - запись либо timeout либо finished в канал
//         }()
//     }

//     for range tasks {
//         result := <-resultChan
//         fmt.Println(result)
//     }

//     fmt.Println("All work completed!")
// }

// func process(task string, result chan<- string, taskDuration int) {
//     fmt.Println("Started", task)
//     timeout := time.After(time.Duration(2.5 * float32(time.Second)))
//     taskDone := time.After(time.Duration(taskDuration) * time.Second)
//     select { //для чтения из нескольких каналов
//     case <-timeout:
//             result <-fmt.Sprint("Timeout", task)
//     case <-taskDone:
//             result <-fmt.Sprint("Finished", task)
//     }
// }


// 7)
// func main() {
//     fmt.Println("Starting work..")

//     tasks := []string{"Task1", "Task2", "Task3"}

//     resultChan := make(chan string, len(tasks))
//     defer close(resultChan)

//     timeout := time.Duration(2.5 * float32(time.Second))

//     ctx, cancel := context.WithTimeout(context.Background(), timeout) //контекст нужен для таймаутов и завершения горутин
//     defer cancel()

//     for i, task := range tasks { 
//         go func() {
//             process(ctx, task, resultChan, i+1)
//         }()
//     }

//     for range tasks {
//         result := <-resultChan
//         fmt.Println(result)
//     }

//     fmt.Println("All work completed!")
// }

// func process(ctx context.Context, task string, result chan<- string, taskDuration int) {
//     fmt.Println("Started", task)

//     taskDone := time.After(time.Duration(taskDuration) * time.Second)
//     select {
//     case <-ctx.Done():
//             result <-fmt.Sprint("Timeout", task)
//     case <-taskDone:
//             result <-fmt.Sprint("Finished", task)
//     }
// }

// 8) mutexes
// 8.1
// type InventoryManager struct {
//     mu sync.RWMutex
//     inventory map[string]int
// }

// func (mim *InventoryManager) Update(name string, val int) {
//     mim.mu.Lock()
//     defer mim.mu.Unlock()
//     mim.inventory[name] = val
// }

// func (mim *InventoryManager) Read(name string) (int, bool) {
//     mim.mu.RLock() //защита на запись с возможностью чттения
//     defer mim.mu.RUnlock()
//     val, ok := mim.inventory[name]
//     return val, ok
// }

// 8.2
// func inventoryManager(ctx context.Context, in <-chan func(map[string]int)) {
//     inventory := map[string]int{}

//     for {
//         select {
//         case <-ctx.Done():
//             return
//         case f := <-in:
//             f(inventory)
//         }
//     }
// }

// type InventoryManager chan func(map[string]int)

// func NewChannelScoreboardManager(ctx context.Context) InventoryManager {
//     ch := make(InventoryManager)
//     go inventoryManager(ctx, ch)
//     return ch
// }

// func (cim InventoryManager) Update(name string, val int) {
//     cim <- func(m map[string]int) {
//         m[name] = val
//     }
// }

// func (cim InventoryManager) Read(name string) (int, bool) {
//     type Result struct {
//         out int
//         ok bool
//     }

//     resultCh := make(chan Result)
//     cim <- func(m map[string]int) {
//         out, ok := m[name]
//         resultCh <- Result{out, ok}
//     }

//     result := <-resultCh
//     return result.out, result.ok
// }
