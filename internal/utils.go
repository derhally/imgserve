package internal

func fileExtFromMimeType(mimeType string) string {
	switch mimeType {
	case MIMEImageJpeg:
		return ".jpeg"
	case MIMEImagePng:
		return ".png"
	default:
		return ""
	}
}
