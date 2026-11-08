package rtnetlink

import "encoding/binary"

// nativeEndian is the byte order of the host, which netlink uses for most
// attribute and header fields.
var nativeEndian binary.ByteOrder = binary.NativeEndian
