package views

func chunks(text string, chunkSize int, maximumChunks int) []string {
	if len(text) == 0 {
		return nil
	}
	if chunkSize >= len(text) {
		return []string{text}
	}
	var chunks []string = make([]string, 0, maximumChunks)
	currentLength := 0
	currentStart := 0
	for i := range text {
		if currentLength == chunkSize {
			chunks = append(chunks, text[currentStart:i])
			currentLength = 0
			currentStart = i
		}
		currentLength++
		if len(chunks) >= maximumChunks-1 {
			break
		}
	}
	chunks = append(chunks, text[currentStart:])
	return chunks
}
