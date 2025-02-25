package main

import (
	"fmt"
	"sync"
	"time"
)

// SomeRepository - интерфейс для получения данных
type SomeRepository interface {
	GetData() string
}

// SomeRepositoryImpl - реализация интерфейса SomeRepository
type SomeRepositoryImpl struct{}

// GetData - метод, который имитирует запрос к базе данных
func (r *SomeRepositoryImpl) GetData() string {
	// Здесь происходит запрос к базе данных
	time.Sleep(2 * time.Second) // Имитация задержки запроса
	return "data from database"
}

// SomeRepositoryProxy - прокси для кэширования данных
type SomeRepositoryProxy struct {
	repository SomeRepository
	cache      map[string]string
	mutex      sync.RWMutex
	ttl        time.Duration
}

// NewSomeRepositoryProxy - конструктор для создания прокси
func NewSomeRepositoryProxy(repository SomeRepository, ttl time.Duration) *SomeRepositoryProxy {
	return &SomeRepositoryProxy{
		repository: repository,
		cache:      make(map[string]string),
		ttl:        ttl,
	}
}

// GetData - метод прокси для получения данных с кэшированием
func (p *SomeRepositoryProxy) GetData() string {
	p.mutex.RLock()
	if data, found := p.cache["data"]; found {
		p.mutex.RUnlock()
		return data // Возвращаем данные из кэша
	}
	p.mutex.RUnlock()

	// Если данных нет в кэше, запрашиваем их у оригинального объекта
	data := p.repository.GetData()

	// Сохраняем данные в кэш с таймером
	p.mutex.Lock()
	p.cache["data"] = data
	go func() {
		time.Sleep(p.ttl)
		p.mutex.Lock()
		delete(p.cache, "data") // Удаляем данные из кэша по истечении времени жизни
		p.mutex.Unlock()
	}()
	p.mutex.Unlock()

	return data
}

func main() {
	repo := &SomeRepositoryImpl{}
	proxy := NewSomeRepositoryProxy(repo, 10*time.Second)

	// Первый запрос - данные будут загружены из базы
	fmt.Println(proxy.GetData())

	// Второй запрос - данные будут возвращены из кэша
	fmt.Println(proxy.GetData())

	// Ждем 11 секунд, чтобы данные из кэша истекли
	time.Sleep(11 * time.Second)

	// Третий запрос - данные снова будут загружены из базы
	fmt.Println(proxy.GetData())
}
