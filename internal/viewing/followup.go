package viewing

import (
	"fmt"
	"strings"
)

// FollowUpMessage drafts a message to the landlord asking only the given
// unresolved questions. It returns "" when there is nothing left to ask.
// It takes questions only, so a user's answers or notes cannot leak into it.
func FollowUpMessage(unresolved []ChecklistItem) string {
	if len(unresolved) == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString("Chào b, mình xem phòng xong và muốn hỏi thêm vài ý:\n")

	for i, it := range unresolved {
		fmt.Fprintf(&b, "%d. %s\n", i+1, it.Question)
	}

	b.WriteString("Cảm ơn b nhiều ạ.")

	return b.String()
}
