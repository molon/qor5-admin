package publish

type List struct {
	PageNumber  int
	Position    int
	ListDeleted bool
	ListUpdated bool
}

func (this List) GetPageNumber() int {
	return this.PageNumber
}

func (this *List) SetPageNumber(pageNumber int) {
	this.PageNumber = pageNumber
}

func (this List) GetPosition() int {
	return this.Position
}

func (this *List) SetPosition(position int) {
	this.Position = position
}

func (this *List) SetListDeleted(listDeleted bool) {
	this.ListDeleted = listDeleted
}

func (this *List) SetListUpdated(listUpdated bool) {
	this.ListUpdated = listUpdated
}
