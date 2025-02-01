package models

import (
	"encoding/xml"
)

type User struct {
	FriendlyName string `xml:"FriendlyName,attr"`
}

type From struct {
	User User `xml:"User"`
}

type To struct {
	User User `xml:"User"`
}

type Text struct {
	Text  string `xml:",chardata"`
	Style string `xml:"Style,attr"`
}

type Message struct {
	Date      string `xml:"Date,attr"`
	Time      string `xml:"Time,attr"`
	DateTime  string `xml:"DateTime,attr"`
	SessionID string `xml:"SessionID,attr"`
	From      From   `xml:"From"`
	To        To     `xml:"To"`
	Text      Text   `xml:"Text"`
}

type Log struct {
	XMLName        xml.Name  `xml:"Log"`
	FirstSessionID string    `xml:"FirstSessionID,attr"`
	LastSessionID  string    `xml:"LastSessionID,attr"`
	Messages       []Message `xml:"Message"`
}
