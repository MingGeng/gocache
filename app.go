import "time"

type Item struct {
	Object interface{}
	Expiration int64
}

func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

const (
    NoExpiration time.Duration = -1
    DefaultExpiration time.Duration = 0
)

type Cache struct {
	defaultExpiration time.Duration
	items map[string]Item
	mu sync.RWMutex
	gcInterval time.Duration
	stopGc chan bool
}

func (c *Cache) gcLoop() {
	ticker := time.NewTicker(c.gcInterval)
    for {
    	select {
    	case <-ticker.C:
    		c.DeleteExpired()
    	case <-c.stopGc:
    		ticker.Stop()
    		return
    	}
    }
}

func (c *Cache) delete(k string) {
	delete(c.items, k)
}

func (c *Cache) DeleteExpired() {
	now := time.Now().UnixNano()
    c.mu.Lock()
    defer c.mu.Unlock()

    for k, v := range c.items {
    	if v.Expiration > 0 && now > v.Expiration {
    		c.delete(k)
    	}
    }   
}

func (c *Cache) Set(k string, v interface{}, d time.Duration) {
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[k] = Item{
		Object: v,
		Expiration: e,
	}
}

func (c *Cache) set(k string, v interface{}, d time.Duration) {
    var e int64
    if d == DefaultExpiration {
        d = c.defaultExpiration
    }
    if d > 0 {
        e = time.Now().Add(d).UnixNano()
    }
    c.items[k] = Item{
        Object:     v,
        Expiration: e,
    }
}

func (c *Cache) get(k string) (interface{}, bool) {
    item, found := c.items[k]
    if !found {
        return nil, false
    }
    if item.Expired() {
        return nil, false
    }
    return item.Object, true
}

func (c *Cache) Add(k string, v interface{}, d time.Duration) error {
    c.mu.Lock()
    _, found := c.get(k)
    if found {
        c.mu.Unlock()
        return fmt.Errorf("Item %s already exists", k)
    }
    c.set(k, v, d)
    c.mu.Unlock()
    return nil
}

func (c *Cache) Get(k string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    item, found := c.items[k]
    if !found {
        return nil, false
    }
    if item.Expired() {
        return nil, false
    }
    return item.Object, true
}













