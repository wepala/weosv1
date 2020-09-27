package entities

//AggregateRoot base struct for microservices to use
type AggregateRoot struct {
	ID         string
	SequenceNo int64
	newEvents  []*Event
}

func (w *AggregateRoot) IsValid() bool {
	return w.ID != ""
}

func (w *AggregateRoot) GetID() string {
	return w.ID
}

func (w *AggregateRoot) NewChange(event *Event) {
	w.newEvents = append(w.newEvents, event)
}

func (w *AggregateRoot) GetNewChanges() []*Event {
	return w.newEvents
}

func (w *AggregateRoot) ApplyEvent(event *Event) {

}

func (w *AggregateRoot) ApplyEventHistory(event []*Event) {
	for _, event := range event {
		w.ApplyEvent(event)
	}
}
