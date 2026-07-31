package permission

// AttachmentInfo is a port of pocketmine\permission\PermissionAttachmentInfo.
type AttachmentInfo struct {
	permission      string
	attachment      *PermissionAttachment
	value           bool
	groupPermission *AttachmentInfo
}

func NewAttachmentInfo(permission string, attachment *PermissionAttachment, value bool, groupPermission *AttachmentInfo) *AttachmentInfo {
	return &AttachmentInfo{permission: permission, attachment: attachment, value: value, groupPermission: groupPermission}
}

func (i *AttachmentInfo) Permission() string                { return i.permission }
func (i *AttachmentInfo) Attachment() *PermissionAttachment { return i.attachment }
func (i *AttachmentInfo) Value() bool                       { return i.value }

// GroupPermissionInfo returns the info of the permission group that caused this permission to be
// set, if any. If nil, the permission was set explicitly (a base permission or attachment entry).
func (i *AttachmentInfo) GroupPermissionInfo() *AttachmentInfo { return i.groupPermission }
