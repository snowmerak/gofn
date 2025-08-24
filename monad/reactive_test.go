package monad

import (
	"sync"
	"testing"
	"time"
)

func TestReactiveBasics(t *testing.T) {
	reactive := NewReactive(42)
	
	// Test Get
	value := reactive.Get()
	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}

	// Test Set
	reactive.Set(100)
	value = reactive.Get()
	if value != 100 {
		t.Errorf("Expected 100, got %d", value)
	}

	// Test Update
	reactive.Update(func(x int) int { return x * 2 })
	value = reactive.Get()
	if value != 200 {
		t.Errorf("Expected 200, got %d", value)
	}
}

func TestReactiveSubscription(t *testing.T) {
	reactive := NewReactive(42)
	
	var receivedValues []int
	var mu sync.Mutex
	
	// Subscribe to changes
	id := reactive.Subscribe(func(oldValue, newValue int) {
		mu.Lock()
		defer mu.Unlock()
		receivedValues = append(receivedValues, newValue)
	})

	// Make some changes
	reactive.Set(100)
	reactive.Set(200)
	reactive.Update(func(x int) int { return x + 50 })

	// Give some time for subscribers to be notified
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	
	if len(receivedValues) != 3 {
		t.Errorf("Expected 3 notifications, got %d", len(receivedValues))
	}
	
	expected := []int{100, 200, 250}
	for i, expected := range expected {
		if i >= len(receivedValues) || receivedValues[i] != expected {
			t.Errorf("Expected %d at index %d, got %v", expected, i, receivedValues)
		}
	}

	// Test unsubscribe
	reactive.Unsubscribe(id)
	receivedValues = nil
	reactive.Set(999)
	
	// Give some time
	time.Sleep(10 * time.Millisecond)
	
	if len(receivedValues) != 0 {
		t.Errorf("Expected no notifications after unsubscribe, got %d", len(receivedValues))
	}
}

func TestReactiveMultipleSubscribers(t *testing.T) {
	reactive := NewReactive(0)
	
	var values1, values2 []int
	var mu1, mu2 sync.Mutex
	
	// Multiple subscribers
	id1 := reactive.Subscribe(func(oldValue, newValue int) {
		mu1.Lock()
		defer mu1.Unlock()
		values1 = append(values1, newValue)
	})
	
	id2 := reactive.Subscribe(func(oldValue, newValue int) {
		mu2.Lock()
		defer mu2.Unlock()
		values2 = append(values2, newValue*2) // Different transformation
	})

	reactive.Set(10)
	reactive.Set(20)

	// Give some time for notifications
	time.Sleep(10 * time.Millisecond)

	mu1.Lock()
	mu2.Lock()
	defer mu1.Unlock()
	defer mu2.Unlock()
	
	if len(values1) != 2 || values1[0] != 10 || values1[1] != 20 {
		t.Errorf("Subscriber 1 got unexpected values: %v", values1)
	}
	
	if len(values2) != 2 || values2[0] != 20 || values2[1] != 40 {
		t.Errorf("Subscriber 2 got unexpected values: %v", values2)
	}

	// Unsubscribe one
	reactive.Unsubscribe(id1)
	values1 = nil
	values2 = nil
	
	reactive.Set(30)
	time.Sleep(10 * time.Millisecond)
	
	if len(values1) != 0 {
		t.Errorf("Unsubscribed subscriber should not receive values: %v", values1)
	}
	
	if len(values2) != 1 || values2[0] != 60 {
		t.Errorf("Active subscriber should receive value: %v", values2)
	}

	reactive.Unsubscribe(id2)
}

func TestMapReactive(t *testing.T) {
	source := NewReactive(10)
	mapped := MapReactive(source, func(x int) string {
		return "value: " + string(rune(x+48))
	})
	
	// Check initial value
	value := mapped.Get()
	expected := "value: :" // 10 + 48 = 58 (':')
	if value != expected {
		t.Errorf("Expected %s, got %s", expected, value)
	}

	// Test that changes propagate
	var receivedValue string
	var mu sync.Mutex
	
	mapped.Subscribe(func(oldVal, newVal string) {
		mu.Lock()
		defer mu.Unlock()
		receivedValue = newVal
	})

	source.Set(42)
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	
	expected = "value: Z" // 42 + 48 = 90 ('Z')
	if receivedValue != expected {
		t.Errorf("Expected %s, got %s", expected, receivedValue)
	}
	
	// Check that mapped reactive also updated
	value = mapped.Get()
	if value != expected {
		t.Errorf("Expected %s, got %s", expected, value)
	}
}

func TestFilterReactive(t *testing.T) {
	source := NewReactive(5)
	filtered := FilterReactive(source, func(x int) bool { return x > 10 })
	
	// Initial value should be zero since 5 <= 10
	value := filtered.Get()
	if value != 0 {
		t.Errorf("Expected 0 (filtered out), got %d", value)
	}

	var receivedValues []int
	var mu sync.Mutex
	
	filtered.Subscribe(func(oldVal, newVal int) {
		mu.Lock()
		defer mu.Unlock()
		receivedValues = append(receivedValues, newVal)
	})

	// This should be filtered out
	source.Set(8)
	time.Sleep(10 * time.Millisecond)

	// This should pass through
	source.Set(15)
	time.Sleep(10 * time.Millisecond)

	// This should be filtered out
	source.Set(3)
	time.Sleep(10 * time.Millisecond)

	// This should pass through
	source.Set(25)
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	
	// Should only have received the values that passed the filter
	expected := []int{15, 25}
	if len(receivedValues) != len(expected) {
		t.Errorf("Expected %d values, got %d: %v", len(expected), len(receivedValues), receivedValues)
	}
	
	for i, exp := range expected {
		if i >= len(receivedValues) || receivedValues[i] != exp {
			t.Errorf("Expected %d at index %d, got %v", exp, i, receivedValues)
		}
	}
	
	// Check final value
	value = filtered.Get()
	if value != 25 {
		t.Errorf("Expected 25, got %d", value)
	}
}

func TestCombineReactives(t *testing.T) {
	r1 := NewReactive(10)
	r2 := NewReactive(20)
	
	combined := CombineReactives(r1, r2, func(a, b int) string {
		return string(rune(a+48)) + "+" + string(rune(b+48))
	})
	
	// Check initial value
	value := combined.Get()
	expected := ":+4" // 10+48=58(':'), 20+48=68('4')
	if value != expected {
		t.Errorf("Expected %s, got %s", expected, value)
	}

	var receivedValues []string
	var mu sync.Mutex
	
	combined.Subscribe(func(oldVal, newVal string) {
		mu.Lock()
		defer mu.Unlock()
		receivedValues = append(receivedValues, newVal)
	})

	// Change first reactive
	r1.Set(30)
	time.Sleep(10 * time.Millisecond)

	// Change second reactive
	r2.Set(40)
	time.Sleep(10 * time.Millisecond)

	// Change both (this should trigger the combine function)
	r1.Set(50)
	time.Sleep(5 * time.Millisecond)
	r2.Set(60)
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	
	// Should have received updates for each change
	if len(receivedValues) < 2 {
		t.Errorf("Expected at least 2 updates, got %d: %v", len(receivedValues), receivedValues)
	}
	
	// Check final value
	finalValue := combined.Get()
	expected = "2<" // 50+48=98('2'), 60+48=108('<')
	if finalValue != expected {
		t.Errorf("Expected %s, got %s", expected, finalValue)
	}
}