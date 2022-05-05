package main

import "fyne.io/fyne/v2/widget"

type titledText struct {
	*widget.RichText
}

func newTitledText(title, text string) titledText {
	titleStyle := widget.RichTextStyleHeading
	titleStyle.TextStyle.Bold = false
	return titledText{
		widget.NewRichText(
			&widget.TextSegment{
				Style: widget.RichTextStyleParagraph,
				Text:  title,
			},
			&widget.TextSegment{
				Style: titleStyle,
				Text:  text,
			},
			&widget.SeparatorSegment{},
		),
	}
}

func (t titledText) SetTitle(s string) {
	t.RichText.Segments[0].(*widget.TextSegment).Text = s
	t.Refresh()
}

func (t titledText) SetBody(s string) {
	t.RichText.Segments[1].(*widget.TextSegment).Text = s
	t.Refresh()
}
