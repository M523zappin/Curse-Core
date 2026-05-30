package adapters

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type FortuneAdapter struct {
	profile gateway.ModelProfile
}

func NewFortune(profile gateway.ModelProfile) *FortuneAdapter {
	return &FortuneAdapter{profile: profile}
}

func (a *FortuneAdapter) Name() string { return "fortune" }
func (a *FortuneAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *FortuneAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	cat := categoryFromQuery(req)
	return &gateway.Response{
		Message: gateway.Message{Role: gateway.RoleAssistant, Content: cat},
		Done:    true,
	}, nil
}

func categoryFromQuery(req *gateway.Prompt) string {
	q := ""
	for _, m := range req.Messages {
		if m.Role == gateway.RoleUser {
			q = strings.ToLower(m.Content)
			break
		}
	}

	switch {
	case strings.Contains(q, "quote") || strings.Contains(q, "wisdom"):
		return randomQuote()
	case strings.Contains(q, "joke") || strings.Contains(q, "humor"):
		return randomJoke()
	case strings.Contains(q, "fact") || strings.Contains(q, "did you know"):
		return randomFact()
	case strings.Contains(q, "motivat") || strings.Contains(q, "inspire"):
		return randomMotivation()
	case strings.Contains(q, "riddle") || strings.Contains(q, "puzzle"):
		return randomRiddle()
	default:
		return randomQuote()
	}
}

var quotes = []string{
	"\"The best way to predict the future is to implement it.\" — David Heinemeier Hansson",
	"\"First, solve the problem. Then, write the code.\" — John Johnson",
	"\"Any fool can write code that a computer can understand. Good programmers write code that humans can understand.\" — Martin Fowler",
	"\"Programs must be written for people to read, and only incidentally for machines to execute.\" — Harold Abelson",
	"\"Debugging is twice as hard as writing the code in the first place. Therefore, if you write the code as cleverly as possible, you are, by definition, not smart enough to debug it.\" — Brian Kernighan",
	"\"Simplicity is the soul of efficiency.\" — Austin Freeman",
	"\"Make it work, make it right, make it fast.\" — Kent Beck",
	"\"The most dangerous phrase in the language is: 'We've always done it this way.'\" — Grace Hopper",
	"\"Talk is cheap. Show me the code.\" — Linus Torvalds",
	"\"Software is a great combination of artistry and engineering.\" — Bill Gates",
	"\"Premature optimization is the root of all evil.\" — Donald Knuth",
	"\"Code is like humor. When you have to explain it, it's bad.\" — Cory House",
	"\"The best programs are written so that computing machines can perform them quickly and so that human beings can understand them clearly.\" — Donald Knuth",
	"\"Sometimes it pays to stay in bed on Monday, rather than spending the rest of the week debugging Monday's code.\" — Dan Salomon",
	"\"A good programmer is someone who always looks both ways before crossing a one-way street.\" — Doug Linder",
}

var jokes = []string{
	"A QA engineer walks into a bar. Orders 1 beer. Orders 999999 beers. Orders 0 beers. Orders -1 beers. Orders a lizard. The real customer walks in and asks where the bathroom is. The bar bursts into flames. QA: \"Yeah, that's what we expected.\"",
	"Why do programmers prefer dark mode? Because light attracts bugs.",
	"Q: How many programmers does it take to change a light bulb? A: None — that's a hardware problem.",
	"I would tell you a UDP joke, but you might not get it.",
	"Q: Why did the Go programmer quit his job? A: He couldn't handle the goroutines.",
	"Q: What's a programmer's favorite hangout? A: The Foo Bar.",
	"A SQL query walks into a bar, walks up to two tables and asks: \"Can I join you?\"",
	"Q: Why do Java programmers wear glasses? A: Because they can't C#.",
	"Q: What do you call a programmer from Finland? A: Nerdic.",
	"Q: Why did the developer go broke? A: Because he used up all his cache.",
}

var facts = []string{
	"The first computer bug was an actual moth found in the Harvard Mark II computer in 1947.",
	"Go was announced in November 2009 and reached version 1.0 in March 2012.",
	"The World Wide Web was invented by Tim Berners-Lee in 1989.",
	"The first computer virus, 'Creeper,' was written in 1971 on ARPANET.",
	"Python's name comes from 'Monty Python's Flying Circus', not the snake.",
	"The TCP/IP protocol suite was developed in the 1970s and adopted as a standard in 1983.",
	"The first 1GB hard drive (1991) weighed about 5 pounds and cost over $1,000.",
	"Git was created by Linus Torvalds in 2005 to manage the Linux kernel development.",
	"The '@' symbol was chosen for email addresses because it was rarely used and made sense as 'at'.",
	"There are over 700 programming languages in active use today.",
}

var motivations = []string{
	"The best time to start is now. The second best time is tomorrow. Don't wait for perfect — ship it.",
	"You don't need to be great to start, but you need to start to be great.",
	"Every expert was once a beginner. Every master was once confused. Keep going.",
	"Done is better than perfect. Perfect never ships. Done learns.",
	"The only way to eat an elephant is one bite at a time. Same goes for refactoring legacy code.",
	"It's not a bug — it's an undocumented feature. And you just documented it.",
	"Your most productive hours are the ones you actually start working. Everything else is just setup.",
	"Code wins arguments. Ship it and let the results speak.",
}

var riddles = []string{
	"Riddle: I speak without a mouth and hear without ears. I have no body, but I come alive with the wind. What am I?\n\nAnswer: An echo.",
	"Riddle: What gets wetter the more it dries?\n\nAnswer: A towel.",
	"Riddle: I have cities, but no houses. I have mountains, but no trees. I have water, but no fish. What am I?\n\nAnswer: A map.",
	"Riddle: The more you take, the more you leave behind. What am I?\n\nAnswer: Footsteps.",
	"Riddle: What can travel around the world while staying in a corner?\n\nAnswer: A stamp.",
}

func randomQuote() string {
	return quotes[rand.Intn(len(quotes))]
}

func randomJoke() string {
	return jokes[rand.Intn(len(jokes))]
}

func randomFact() string {
	return facts[rand.Intn(len(facts))]
}

func randomMotivation() string {
	return motivations[rand.Intn(len(motivations))]
}

func randomRiddle() string {
	return riddles[rand.Intn(len(riddles))]
}
