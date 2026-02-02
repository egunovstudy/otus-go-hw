package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v any) *ListItem
	PushBack(v any) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len   int
	front *ListItem
	back  *ListItem
}

func NewList() List {
	return &list{}
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.front
}

func (l *list) Back() *ListItem {
	return l.back
}

func (l *list) PushFront(v any) *ListItem {
	item := &ListItem{Value: v}

	if l.len == 0 {
		l.front = item
		l.back = item
		l.len = 1
		return item
	}

	item.Next = l.front
	l.front.Prev = item
	l.front = item
	l.len++
	return item
}

func (l *list) PushBack(v any) *ListItem {
	item := &ListItem{Value: v}

	if l.len == 0 {
		l.front = item
		l.back = item
		l.len = 1
		return item
	}

	item.Prev = l.back
	l.back.Next = item
	l.back = item
	l.len++
	return item
}

func (l *list) Remove(i *ListItem) {
	// По условию i всегда из этого списка.

	// левый сосед
	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		// удаляем front
		l.front = i.Next
	}

	// правый сосед
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		// удаляем back
		l.back = i.Prev
	}

	// отвязываем
	i.Prev = nil
	i.Next = nil

	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	// По условию i всегда из этого списка.
	if l.len <= 1 || i == l.front {
		return
	}

	// 1) вырезаем i из текущей позиции
	if i.Prev != nil {
		i.Prev.Next = i.Next
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		// i был back
		l.back = i.Prev
	}

	// 2) вставляем в начало
	i.Prev = nil
	i.Next = l.front
	l.front.Prev = i
	l.front = i
}
