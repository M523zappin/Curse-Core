package gateway

type ContextTruncator struct {
	profile ModelProfile
}

func NewContextTruncator(profile ModelProfile) *ContextTruncator {
	return &ContextTruncator{profile: profile}
}

func (ct *ContextTruncator) Truncate(msgs []Message) []Message {
	total := countTokens(msgs)
	if total <= ct.profile.ContextWindow {
		return msgs
	}
	var kept []Message
	if len(msgs) > 0 && msgs[0].Role == RoleSystem {
		kept = append(kept, msgs[0])
		msgs = msgs[1:]
	}
	budget := ct.profile.ContextWindow - countTokens(kept)
	for i := len(msgs) - 1; i >= 0 && budget > 0; i-- {
		tokens := msgTokens(msgs[i])
		if tokens <= budget {
			kept = append([]Message{msgs[i]}, kept...)
			budget -= tokens
		}
	}
	if countTokens(kept) >= ct.profile.ContextWindow && len(kept) > 0 {
		kept = kept[1:]
	}
	return kept
}

func countTokens(msgs []Message) int {
	total := 0
	for _, m := range msgs {
		total += msgTokens(m)
	}
	return total
}

func msgTokens(m Message) int {
	return len(m.Content) / 4
}

func (ct *ContextTruncator) Profile() ModelProfile {
	return ct.profile
}
