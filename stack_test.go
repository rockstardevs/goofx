package goofx_test

import (
	"encoding/xml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rockstardevs/goofx"
)

var _ = Describe("goofx", func() {
	Describe("NewStack()", func() {
		It("should return an initialized empty stack", func() {
			s := goofx.NewStack()
			Expect(s).ToNot(BeNil())
			Expect(s.IsEmpty()).To(BeTrue())
			Expect(s.Size()).To(Equal(0))
		})
	})
	Describe("TagStack", func() {
		var s goofx.TagStack
		BeforeEach(func() {
			s = goofx.NewStack()
		})
		Describe("Push()", func() {
			It("should add the given element to the stack", func() {
				t := xml.StartElement{}
				s.Push(&t)
				Expect(s.IsEmpty()).To(BeFalse())
				Expect(s.Size()).To(Equal(1))
			})
		})
		Describe("Pop()", func() {
			It("should remove the last element from the stack", func() {
				t1 := xml.StartElement{Name: xml.Name{Local: "test1"}}
				t2 := xml.StartElement{Name: xml.Name{Local: "test2"}}
				s.Push(&t1)
				s.Push(&t2)
				Expect(s.IsEmpty()).To(BeFalse())
				Expect(s.Size()).To(Equal(2))
				t, err := s.Pop()
				Expect(err).To(BeNil())
				Expect(t).To(Equal(&t2))
				Expect(s.IsEmpty()).To(BeFalse())
				Expect(s.Size()).To(Equal(1))
			})
			It("should return an error when popping an empty stack", func() {
				t, err := s.Pop()
				Expect(err).To(MatchError("error - popping from empty stack"))
				Expect(t).To(BeNil())
			})
		})
	})
})
