// Copyright 2012-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkgs

// PkgDeps defines the expected dependencies between packages in
// the Go source tree. It is a statement of policy.
// Changes should not be made to this map without prior discussion.
//
// The map contains two kinds of entries:
// 1) Lower-case keys are standard import paths and list the
// allowed imports in that package.
// 2) Upper-case keys define aliases for package sets, which can then
// be used as dependencies by other rules.
//
var PkgDeps = map[string][]string{
	"builtin": {}, // dud package

	// L0 is the lowest level, core, nearly unavoidable packages.
	"unsafe":                  {},
	"errors":                  {},
	"runtime/internal/sys":    {},
	"runtime/internal/atomic": {"unsafe", "runtime/internal/sys"},
	"runtime":                 {"unsafe", "runtime/internal/atomic", "runtime/internal/sys"},
	"internal/race":           {"runtime", "unsafe"},
	"internal/cpu":            {"runtime"},
	"sync/atomic":             {"unsafe"},
	"sync":                    {"runtime", "unsafe", "internal/race", "sync/atomic"},
	"io":                      {"errors", "sync"},

	"L0": {
		"unsafe",
		"errors",
		"runtime/internal/atomic",
		"runtime",
		"internal/cpu",
		"sync/atomic",
		"sync",
		"io",
	},

	"LX": { //#### Added this set because these items are nowhere else
		"runtime/internal/sys",
		"internal/race",
	},

	//--------------------------------------------------------------------------------
	// L1 adds simple functions and strings processing,
	// but not Unicode tables.
	"math/bits":     {},
	"math":          {"internal/cpu", "unsafe"},
	"math/cmplx":    {"math"},
	"math/rand":     {"math" /*L0*/, "sync"},
	"unicode/utf16": {},
	"unicode/utf8":  {},
	"strconv":       {"unicode/utf8", "math" /*L0*/, "errors"},

	"L1": {
		"L0",
		"math/bits",
		"math",
		"math/cmplx",
		"math/rand",
		"unicode/utf16",
		"unicode/utf8",
		"strconv",
		//"sort", //#### Removed because it uses L3's "reflect"
	},

	//--------------------------------------------------------------------------------
	// L2 adds Unicode and strings processing.
	"unicode": {},
	"bytes":   {"unicode", "unicode/utf8" /*L0*/, "io", "errors", "internal/cpu"},
	"bufio":   {"unicode/utf8", "bytes" /*L0*/, "io", "errors"},
	"strings": {"unicode", "unicode/utf8" /*L0*/, "io", "errors", "internal/cpu"},
	"path":    {"unicode/utf8", "strings" /*L0*/, "errors"},

	"L2": {
		"L1",
		"unicode",
		"bytes",
		"bufio",
		"strings",
		"path",
		"sort", //#### Moved from L1; listed further on -- Error?
	},

	//--------------------------------------------------------------------------------
	// L3 adds reflection and some basic utility packages
	// and interface definitions, but nothing that makes
	// system calls.
	"hash":         { /*L2*/ "io"}, // interfaces
	"hash/adler32": {"hash" /*L2*/},
	"hash/crc32":   {"hash" /*L2*/, "internal/cpu", "sync", "unsafe"},
	"hash/crc64":   {"hash" /*L2*/},
	"hash/fnv":     {"hash" /*L2*/},

	"crypto/internal/cipherhw": {},                               //#### Added
	"crypto":                   {"hash" /*L2*/, "io", "strconv"}, // interfaces
	"crypto/subtle":            {},
	"crypto/cipher":            {"crypto/subtle" /*L2*/, "unsafe", "runtime", "io", "errors"},

	"image/color":         { /*L2*/ }, // interfaces
	"image/color/palette": {"image/color" /*L2*/},
	"image":               {"image/color" /*L2*/, "errors", "strconv", "bufio", "io"}, // interfaces

	"reflect": { /*L2*/ "runtime", "sync", "math", "strconv", "unicode", "unicode/utf8", "unsafe"},

	"encoding/base32": { /*L2*/ "io", "strings", "bytes", "strconv"},
	"encoding/base64": { /*L2*/ "io", "strconv"},
	"encoding/binary": {"reflect" /*L2*/, "io", "errors", "math"},

	"L3": {
		"L2",
		"hash",
		"hash/adler32",
		"hash/crc32",
		"hash/crc64",
		"hash/fnv",
		"crypto",
		"crypto/cipher",
		"crypto/subtle",
		"crypto/internal/cipherhw",
		"image/color",
		"image",
		"image/color/palette",
		"reflect",
		"encoding/base32",
		"encoding/base64",
		"encoding/binary",
	},

	// End of linear dependency definitions.
	//--------------------------------------------------------------------------------
	"sort": {"reflect"},

	//--------------------------------------------------------------------------------
	// Operating system access.
	"internal/syscall/windows/sysdll": {}, //#### Added
	"syscall": {"internal/race", "internal/syscall/windows/sysdll", "unicode/utf16",
		/*L0*/ "sync", "runtime", "unsafe", "errors", "io", "sync/atomic"},
	"internal/syscall/unix":             {"syscall" /*L0*/, "unsafe", "sync/atomic"},
	"internal/syscall/windows":          {"internal/syscall/windows/sysdll", "syscall" /*L0*/, "unsafe"},
	"internal/syscall/windows/registry": {"syscall", "internal/syscall/windows/sysdll", "unicode/utf16" /*L0*/, "io", "errors", "unsafe"},
	"time": {
		// "L0" without the "io" package:
		"errors",
		"runtime",
		//NOT:"runtime/internal/atomic",
		"sync",
		//NOT:"sync/atomic",
		//NOT:"unsafe",

		// Other time dependencies:
		"syscall",
		"internal/syscall/windows/registry",
	},

	"internal/poll": {"internal/race", "syscall", "time", "unicode/utf16", "unicode/utf8",
		/*L0*/ "errors", "io", "sync", "sync/atomic", "unsafe", "runtime"},

	"os": {"syscall", "time", "internal/poll", "internal/syscall/windows",
		/*L1*/ "sync/atomic", "sync", "errors", "io", "runtime", "unsafe", "unicode/utf16"},
	"fmt": {"os", "reflect" /*L1*/, "io", "errors", "math", "strconv", "sync", "unicode/utf8"},
	// Formatted I/O: few dependencies (L1) but we must add reflect.
	"context": {"errors", "fmt", "reflect", "sync", "time"},
	"math/big": {"fmt" /*L2*/, "encoding/binary", "bytes", "strings",
		"errors", "sync", "io", "math", "math/bits", "math/rand", "strconv"},

	"os/signal":     {"os", "syscall" /*L2*/, "sync"},
	"path/filepath": {"os", "syscall", "sort" /*L2*/, "errors", "runtime", "strings", "unicode/utf8"}, //#### Added: "sort"
	"io/ioutil":     {"os", "path/filepath", "time", "sort" /*L2*/, "sync", "io", "bytes", "strconv"}, //#### Added: "sort"
	"os/exec":       {"os", "context", "path/filepath", "syscall" /*L2*/, "runtime", "strings", "bytes", "strconv", "errors", "io", "sync"},

	// OS enables basic operating system functionality,
	// but not direct use of package syscall, nor os/signal.
	"OS": {
		"time",
		"os",
		"path/filepath",
		"io/ioutil",
		"os/exec",
	},

	//--------------------------------------------------------------------------------
	"log": {"os", "fmt", "time" /*L1*/, "sync", "io", "runtime"},

	// L4 is defined as L3+fmt+log+time, because in general once
	// you're using L3 packages, use of fmt, log, or time is not a big deal.
	"L4": {
		"time",
		"L3",
		"fmt",
		"log",
	},
	"flag": { /*L4,OS*/ "time", "errors", "os", "io", "fmt", "strconv", "reflect", "sort"},

	// Packages used by testing must be low-level (L2+fmt).
	"regexp/syntax": {"sort" /*L2*/, "bytes", "unicode", "strings", "unicode/utf8", "strconv"},                                //#### Added: "sort"
	"regexp":        {"regexp/syntax", "sort" /*L2*/, "sync", "strconv", "io", "bytes", "unicode", "unicode/utf8", "strings"}, //#### Added: "sort"

	"runtime/debug": {"fmt", "os", "time", "sort" /*L2*/, "runtime"}, //#### Added: "sort"

	"runtime/trace":  { /*L0*/ "io", "runtime"},
	"text/tabwriter": { /*L2*/ "io", "bytes", "unicode/utf8"},
	"runtime/pprof": {"compress/gzip" /*below*/, "context", "encoding/binary", "fmt", "io/ioutil", "os", "text/tabwriter", "time", "sort",
		/*L2*/ "sync", "math", "io", "strings", "bufio", "bytes", "strconv", "errors", "runtime", "unsafe"}, //#### Added: "sort"

	"testing": {"flag", "fmt", "internal/race", "runtime/debug", "runtime/trace", "time", "sort",
		/*L2*/ "sync/atomic", "bytes", "strings", "sort", "errors", "os", "io", "sync", "runtime", "strconv"}, //#### Added: "sort"
	"testing/iotest": {"log" /*L2*/, "io", "errors"},
	"testing/quick":  {"flag", "fmt", "reflect", "time" /*L2*/, "math/rand", "math", "strings"},
	"internal/testenv": {"flag", "testing", "syscall",
		/*L2/OS*/ "errors", "strings", "path/filepath", "os/exec", "strconv", "os", "runtime", "io/ioutil", "sync"},

	//--------------------------------------------------------------------------------
	//--------------------------------------------------------------------------------
	//--------------------------------------------------------------------------------
	// Go parser.
	"go/token":   {"L4"},
	"go/ast":     {"L4", "OS", "go/scanner", "go/token"},
	"go/doc":     {"L4", "go/ast", "go/token", "regexp", "text/template"},
	"go/parser":  {"L4", "OS", "go/ast", "go/scanner", "go/token"},
	"go/printer": {"L4", "OS", "go/ast", "go/scanner", "go/token", "text/tabwriter"},
	"go/scanner": {"L4", "OS", "go/token"},

	"GOPARSER": {
		"go/ast",
		"go/doc",
		"go/parser",
		"go/printer",
		"go/scanner",
		"go/token",
	},

	"go/format":       {"L4", "GOPARSER", "internal/format"},
	"internal/format": {"L4", "GOPARSER"},

	// Go type checking.
	"go/constant":               {"L4", "go/token", "math/big"},
	"go/importer":               {"L4", "go/build", "go/internal/gccgoimporter", "go/internal/gcimporter", "go/internal/srcimporter", "go/token", "go/types"},
	"go/internal/gcimporter":    {"L4", "OS", "go/build", "go/constant", "go/token", "go/types", "text/scanner"},
	"go/internal/gccgoimporter": {"L4", "OS", "debug/elf", "go/constant", "go/token", "go/types", "text/scanner"},
	"go/internal/srcimporter":   {"L4", "fmt", "go/ast", "go/build", "go/parser", "go/token", "go/types", "path/filepath"},
	"go/types":                  {"L4", "GOPARSER", "container/heap", "go/constant"},

	//--------------------------------------------------------------------------------
	// One of a kind.
	"container/list": {}, //#### Added.
	"container/ring": {}, //#### Added.
	"container/heap": {"sort"},

	"internal/singleflight": {"sync"},

	"archive/tar":              {"L4", "OS", "syscall"},
	"archive/zip":              {"L4", "OS", "compress/flate"},
	"compress/bzip2":           {"L4"},
	"compress/flate":           {"L4"},
	"compress/gzip":            {"L4", "compress/flate"},
	"compress/lzw":             {"L4"},
	"compress/zlib":            {"L4", "compress/flate"},
	"database/sql":             {"L4", "container/list", "context", "database/sql/driver", "database/sql/internal"},
	"database/sql/driver":      {"L4", "context", "time", "database/sql/internal"},
	"debug/dwarf":              {"L4"},
	"debug/elf":                {"L4", "OS", "debug/dwarf", "compress/zlib"},
	"debug/gosym":              {"L4"},
	"debug/macho":              {"L4", "OS", "debug/dwarf"},
	"debug/pe":                 {"L4", "OS", "debug/dwarf"},
	"debug/plan9obj":           {"L4", "OS"},
	"encoding":                 {"L4"},
	"encoding/ascii85":         {"L4"},
	"encoding/asn1":            {"L4", "math/big"},
	"encoding/csv":             {"L4"},
	"encoding/gob":             {"L4", "OS", "encoding"},
	"encoding/hex":             {"L4"},
	"encoding/json":            {"L4", "encoding"},
	"encoding/pem":             {"L4"},
	"encoding/xml":             {"L4", "encoding"},
	"go/build":                 {"L4", "OS", "GOPARSER"},
	"html":                     {"L4"},
	"image/draw":               {"L4", "image/internal/imageutil"},
	"image/gif":                {"L4", "compress/lzw", "image/color/palette", "image/draw"},
	"image/internal/imageutil": {"L4"},
	"image/jpeg":               {"L4", "image/internal/imageutil"},
	"image/png":                {"L4", "compress/zlib"},
	"index/suffixarray":        {"L4", "regexp"},
	"internal/trace":           {"L4", "OS"},
	"mime":                     {"L4", "OS", "syscall", "internal/syscall/windows/registry"},
	"mime/quotedprintable":     {"L4"},
	"net/internal/socktest":    {"L4", "OS", "syscall"},
	"net/url":                  {"L4"},
	"plugin":                   {"L0", "OS", "CGO"},
	"runtime/pprof/internal/profile": {"L4", "OS", "compress/gzip", "regexp"},
	"testing/internal/testdeps":      {"L4", "runtime/pprof", "regexp"},
	"text/scanner":                   {"L4", "OS"},
	"text/template/parse":            {"L4"},

	"html/template": {
		"L4", "OS", "encoding/json", "html", "text/template",
		"text/template/parse",
	},
	"text/template": {
		"L4", "OS", "net/url", "text/template/parse",
	},

	// Fake entry to satisfy the pseudo-import "C"
	// that shows up in programs that use cgo.
	"C": {},

	// Cgo.
	// If you add a dependency on CGO, you must add the package to
	// cgoPackages in cmd/dist/test.go.
	"runtime/cgo": {"L0", "C"},
	"CGO":         {"C", "runtime/cgo"},

	// Race detector/MSan uses cgo.
	"runtime/race": {"C"},
	"runtime/msan": {"C"},

	// Plan 9 alone needs io/ioutil and os.
	"os/user": {"L4", "CGO", "io/ioutil", "os", "syscall"},

	"internal/nettrace": {}, //#### Added.

	// Basic networking.
	// Because net must be used by any package that wants to
	// do networking portably, it must have a small dependency set: just L0+basic os.
	"net": {
		"L0", "CGO",
		"context", "math/rand", "os", "reflect", "sort", "syscall", "time",
		"internal/nettrace", "internal/poll",
		"internal/syscall/windows", "internal/singleflight", "internal/race",
		"golang_org/x/net/lif", "golang_org/x/net/route",
	},

	// NET enables use of basic network-related packages.
	"NET": {
		"net",
		"mime",
		"net/textproto",
		"net/url",
	},

	// Uses of networking.
	"log/syslog":    {"L4", "OS", "net"},
	"net/mail":      {"L4", "NET", "OS", "mime"},
	"net/textproto": {"L4", "OS", "net"},

	// Core crypto.
	"crypto/aes":    {"L3"},
	"crypto/des":    {"L3"},
	"crypto/hmac":   {"L3"},
	"crypto/md5":    {"L3"},
	"crypto/rc4":    {"L3"},
	"crypto/sha1":   {"L3"},
	"crypto/sha256": {"L3"},
	"crypto/sha512": {"L3"},

	"CRYPTO": {
		"crypto/aes",
		"crypto/des",
		"crypto/hmac",
		"crypto/md5",
		"crypto/rc4",
		"crypto/sha1",
		"crypto/sha256",
		"crypto/sha512",
		"golang_org/x/crypto/chacha20poly1305",
		"golang_org/x/crypto/curve25519",
		"golang_org/x/crypto/poly1305",
	},

	// Random byte, number generation.
	// This would be part of core crypto except that it imports
	// math/big, which imports fmt.
	"crypto/rand": {"L4", "CRYPTO", "OS", "math/big", "syscall", "internal/syscall/unix"},

	// Mathematical crypto: dependencies on fmt (L4) and math/big.
	// We could avoid some of the fmt, but math/big imports fmt anyway.
	"crypto/dsa":      {"L4", "CRYPTO", "math/big"},
	"crypto/ecdsa":    {"L4", "CRYPTO", "crypto/elliptic", "math/big", "encoding/asn1"},
	"crypto/elliptic": {"L4", "CRYPTO", "math/big"},
	"crypto/rsa":      {"L4", "CRYPTO", "crypto/rand", "math/big"},

	"CRYPTO-MATH": {
		"CRYPTO",
		"crypto/dsa",
		"crypto/ecdsa",
		"crypto/elliptic",
		"crypto/rand",
		"crypto/rsa",
		"encoding/asn1",
		"math/big", //TODO: put this somewhere better
	},

	// SSL/TLS.
	"crypto/tls": {
		"L4", "CRYPTO-MATH", "OS",
		"container/list", "crypto/x509", "encoding/pem", "net", "syscall",
	},
	"crypto/x509": {
		"L4", "CRYPTO-MATH", "OS", "CGO",
		"crypto/x509/pkix", "encoding/pem", "encoding/hex", "net", "os/user", "syscall",
	},
	"crypto/x509/pkix": {"L4", "CRYPTO-MATH"},

	// Simple net+crypto-aware packages.
	"mime/multipart": {"L4", "OS", "mime", "crypto/rand", "net/textproto", "mime/quotedprintable"},
	"net/smtp":       {"L4", "CRYPTO", "NET", "crypto/tls"},

	// HTTP, kingpin of dependencies.
	"net/http": {
		"L4", "NET", "OS",
		"compress/gzip",
		"container/list",
		"context",
		"crypto/rand",
		"crypto/tls",
		"golang_org/x/net/http2/hpack",
		"golang_org/x/net/idna",
		"golang_org/x/net/lex/httplex",
		"golang_org/x/net/proxy",
		"golang_org/x/text/unicode/norm",
		"golang_org/x/text/width",
		"internal/nettrace",
		"mime/multipart",
		"net/http/httptrace",
		"net/http/internal",
		"runtime/debug",
	},
	"net/http/internal":  {"L4"},
	"net/http/httptrace": {"context", "crypto/tls", "internal/nettrace", "net", "reflect", "time"},

	// HTTP-using packages.
	"expvar":             {"L4", "OS", "encoding/json", "net/http"},
	"net/http/cgi":       {"L4", "NET", "OS", "crypto/tls", "net/http", "regexp"},
	"net/http/cookiejar": {"L4", "NET", "net/http"},
	"net/http/fcgi":      {"L4", "NET", "OS", "context", "net/http", "net/http/cgi"},
	"net/http/httptest":  {"L4", "NET", "OS", "crypto/tls", "flag", "net/http", "net/http/internal", "crypto/x509"},
	"net/http/httputil":  {"L4", "NET", "OS", "context", "net/http", "net/http/internal"},
	"net/http/pprof":     {"L4", "OS", "html/template", "net/http", "runtime/pprof", "runtime/trace"},
	"net/rpc":            {"L4", "NET", "encoding/gob", "html/template", "net/http"},
	"net/rpc/jsonrpc":    {"L4", "NET", "encoding/json", "net/rpc"},
}

//================================================================================
