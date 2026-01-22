package loadbalancer

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// Single valid PEM certificate without extra whitespace
	testCertificateClean = `-----BEGIN CERTIFICATE-----
MIIFazCCA1OgAwIBAgIUG4u+4CfiGCzPH98t08AxyVA4C5gwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMjAzMTcxMzE1MDhaFw0yMzAz
MTcxMzE1MDhaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggIiMA0GCSqGSIb3DQEB
AQUAA4ICDwAwggIKAoICAQDvZxn/+zTyW4ERvKkwWomUpi8o2tJz1dver+DJrK3g
CNlUoYzR29CWs7jK81XMsvmqL5MzTwP7Hsmd1q69FI+WcPE1aab9909IAk/GGiK2
zTleN3EQQpXn7tnyPtZaT8Y13yFHpC5RgQzTE8CR9Zi2OLEyxGQ36pA6190nxVg2
SLlafNGTZtJvN1/7r9mJhElbrUE+joeXLwNojH/nXk5W/XwzZboIHsSFVYjK2zlg
oqC6+Apo9xF9oY7nlAhQ0KKWvQVbwjGPmVN17ETodHsKJZBoXxDsZUTGCDDJCimz
W64a79lWIxiyOQ4/7Tn2FhPY0omH5UbWCPq2MnXebkOjgcvUY4NIwzpYV0W2GGGG
wwni9fllYAM9OD3bvPX5OaAWNJT6r6EassjWlv0TeGxE+BZZ+g3xPqHTGz2wHzC5
5anLDjM+4vsBVkfkV3SY7W83m74VQ+Qa3WaNXzinL0keFxupLXY8bKhEzXzSLKz+
Br8PGeGRgaSDd1kpFPg+ujN8qvgo0RDI8IqLS7zbXFoQrp1x/dWm9S8DVEXVoUA1
WQnVuWECoBS4Zf41d04pfdCttnN9zHosgwXbJ8m0TfvfkuhViuVAN/p+Mp9WnQ+H
11HEnWpSfOh7ZPjTzjuAsevVfV4g4a3kczMv1rpNPzUUPwtAqx9239wwIR9Y15cl
OQIDAQABo1MwUTAdBgNVHQ4EFgQUXIqlbj5uMUFT/jqM2kWvYZtDNkcwHwYDVR0j
BBgwFoAUXIqlbj5uMUFT/jqM2kWvYZtDNkcwDwYDVR0TAQH/BAUwAwEB/zANBgkq
hkiG9w0BAQsFAAOCAgEACM+xbb9okUGOqkQZkgtxwyQtzotXTW6yHbgfxhBwwuv+
sImOT29L9XUYnS9q8+BKAHuFkrSEIpGFUUhCUfSkle6fnGG5oNOZmCtK37thfQU9
v4Bo8qBZHjDHwk9VTtaZmAk0KbxfhugyWVCVlmDZroSCOiWGkTVhsXaKDkbw4El0
2scygbACtVxnE5Z9fSAw1OPYrYaG2InGL4/0uRez8iyuPOe5Cb/Dd49uxqsGQd3R
C7J4/oZpvoEzQRmjLboQsC90SfjhSiphGBSbaJBddl00k5VsUrqKXZSM8qQqUfV/
nlBmb2NnUlkde8KHs0PjhBho+/7fj+L7mFa2l5jfuiltuqvh2ZyZtRcwgbvbeiLO
fPIiLCgSns0j+Y2EkKWkEJzErPVnl97ZBKYrPZbdX0V9ogoL/jxEy79lo9Js29v6
RF66goJU02EJe502i7Xrs31YCKnHgvz50L6akBiadR6kkMuWvBuwYz0IZKTLqxjd
08GeRBUyalPVhtfJo3MutnaIe/ZVU7KAIwKVgom3KODcTiZYPWtV+1g/E/7p5hhv
BDG1jrIQsVkdn25fa5sd5OPkP/l0Quv5zmzPI7KS+KfeY/v4qA90m4i6vFNFTmm0
HSWWBYNTxnR1b96PITrts8MyjoXLX6Bu1VFNJPr2Jg02LVVoq6RIbe1Uo619ojA=
-----END CERTIFICATE-----
`

	// Single PEM certificate with extra whitespace
	testCertificateWithWhitespace = `

-----BEGIN CERTIFICATE-----

MIIFazCCA1OgAwIBAgIUG4u+4CfiGCzPH98t08AxyVA4C5gwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMjAzMTcxMzE1MDhaFw0yMzAz
MTcxMzE1MDhaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggIiMA0GCSqGSIb3DQEB
AQUAA4ICDwAwggIKAoICAQDvZxn/+zTyW4ERvKkwWomUpi8o2tJz1dver+DJrK3g
CNlUoYzR29CWs7jK81XMsvmqL5MzTwP7Hsmd1q69FI+WcPE1aab9909IAk/GGiK2
zTleN3EQQpXn7tnyPtZaT8Y13yFHpC5RgQzTE8CR9Zi2OLEyxGQ36pA6190nxVg2
SLlafNGTZtJvN1/7r9mJhElbrUE+joeXLwNojH/nXk5W/XwzZboIHsSFVYjK2zlg
oqC6+Apo9xF9oY7nlAhQ0KKWvQVbwjGPmVN17ETodHsKJZBoXxDsZUTGCDDJCimz
W64a79lWIxiyOQ4/7Tn2FhPY0omH5UbWCPq2MnXebkOjgcvUY4NIwzpYV0W2GGGG
wwni9fllYAM9OD3bvPX5OaAWNJT6r6EassjWlv0TeGxE+BZZ+g3xPqHTGz2wHzC5
5anLDjM+4vsBVkfkV3SY7W83m74VQ+Qa3WaNXzinL0keFxupLXY8bKhEzXzSLKz+
Br8PGeGRgaSDd1kpFPg+ujN8qvgo0RDI8IqLS7zbXFoQrp1x/dWm9S8DVEXVoUA1
WQnVuWECoBS4Zf41d04pfdCttnN9zHosgwXbJ8m0TfvfkuhViuVAN/p+Mp9WnQ+H
11HEnWpSfOh7ZPjTzjuAsevVfV4g4a3kczMv1rpNPzUUPwtAqx9239wwIR9Y15cl
OQIDAQABo1MwUTAdBgNVHQ4EFgQUXIqlbj5uMUFT/jqM2kWvYZtDNkcwHwYDVR0j
BBgwFoAUXIqlbj5uMUFT/jqM2kWvYZtDNkcwDwYDVR0TAQH/BAUwAwEB/zANBgkq
hkiG9w0BAQsFAAOCAgEACM+xbb9okUGOqkQZkgtxwyQtzotXTW6yHbgfxhBwwuv+
sImOT29L9XUYnS9q8+BKAHuFkrSEIpGFUUhCUfSkle6fnGG5oNOZmCtK37thfQU9
v4Bo8qBZHjDHwk9VTtaZmAk0KbxfhugyWVCVlmDZroSCOiWGkTVhsXaKDkbw4El0
2scygbACtVxnE5Z9fSAw1OPYrYaG2InGL4/0uRez8iyuPOe5Cb/Dd49uxqsGQd3R
C7J4/oZpvoEzQRmjLboQsC90SfjhSiphGBSbaJBddl00k5VsUrqKXZSM8qQqUfV/
nlBmb2NnUlkde8KHs0PjhBho+/7fj+L7mFa2l5jfuiltuqvh2ZyZtRcwgbvbeiLO
fPIiLCgSns0j+Y2EkKWkEJzErPVnl97ZBKYrPZbdX0V9ogoL/jxEy79lo9Js29v6
RF66goJU02EJe502i7Xrs31YCKnHgvz50L6akBiadR6kkMuWvBuwYz0IZKTLqxjd
08GeRBUyalPVhtfJo3MutnaIe/ZVU7KAIwKVgom3KODcTiZYPWtV+1g/E/7p5hhv
BDG1jrIQsVkdn25fa5sd5OPkP/l0Quv5zmzPI7KS+KfeY/v4qA90m4i6vFNFTmm0
HSWWBYNTxnR1b96PITrts8MyjoXLX6Bu1VFNJPr2Jg02LVVoq6RIbe1Uo619ojA=

-----END CERTIFICATE-----

`

	// Certificate chain with whitespace and comments between certificates
	testCertificateChainWithComments = `
comment before first cert

-----BEGIN CERTIFICATE-----
MIIFazCCA1OgAwIBAgIUG4u+4CfiGCzPH98t08AxyVA4C5gwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMjAzMTcxMzE1MDhaFw0yMzAz
MTcxMzE1MDhaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggIiMA0GCSqGSIb3DQEB
AQUAA4ICDwAwggIKAoICAQDvZxn/+zTyW4ERvKkwWomUpi8o2tJz1dver+DJrK3g
CNlUoYzR29CWs7jK81XMsvmqL5MzTwP7Hsmd1q69FI+WcPE1aab9909IAk/GGiK2
zTleN3EQQpXn7tnyPtZaT8Y13yFHpC5RgQzTE8CR9Zi2OLEyxGQ36pA6190nxVg2
SLlafNGTZtJvN1/7r9mJhElbrUE+joeXLwNojH/nXk5W/XwzZboIHsSFVYjK2zlg
oqC6+Apo9xF9oY7nlAhQ0KKWvQVbwjGPmVN17ETodHsKJZBoXxDsZUTGCDDJCimz
W64a79lWIxiyOQ4/7Tn2FhPY0omH5UbWCPq2MnXebkOjgcvUY4NIwzpYV0W2GGGG
wwni9fllYAM9OD3bvPX5OaAWNJT6r6EassjWlv0TeGxE+BZZ+g3xPqHTGz2wHzC5
5anLDjM+4vsBVkfkV3SY7W83m74VQ+Qa3WaNXzinL0keFxupLXY8bKhEzXzSLKz+
Br8PGeGRgaSDd1kpFPg+ujN8qvgo0RDI8IqLS7zbXFoQrp1x/dWm9S8DVEXVoUA1
WQnVuWECoBS4Zf41d04pfdCttnN9zHosgwXbJ8m0TfvfkuhViuVAN/p+Mp9WnQ+H
11HEnWpSfOh7ZPjTzjuAsevVfV4g4a3kczMv1rpNPzUUPwtAqx9239wwIR9Y15cl
OQIDAQABo1MwUTAdBgNVHQ4EFgQUXIqlbj5uMUFT/jqM2kWvYZtDNkcwHwYDVR0j
BBgwFoAUXIqlbj5uMUFT/jqM2kWvYZtDNkcwDwYDVR0TAQH/BAUwAwEB/zANBgkq
hkiG9w0BAQsFAAOCAgEACM+xbb9okUGOqkQZkgtxwyQtzotXTW6yHbgfxhBwwuv+
sImOT29L9XUYnS9q8+BKAHuFkrSEIpGFUUhCUfSkle6fnGG5oNOZmCtK37thfQU9
v4Bo8qBZHjDHwk9VTtaZmAk0KbxfhugyWVCVlmDZroSCOiWGkTVhsXaKDkbw4El0
2scygbACtVxnE5Z9fSAw1OPYrYaG2InGL4/0uRez8iyuPOe5Cb/Dd49uxqsGQd3R
C7J4/oZpvoEzQRmjLboQsC90SfjhSiphGBSbaJBddl00k5VsUrqKXZSM8qQqUfV/
nlBmb2NnUlkde8KHs0PjhBho+/7fj+L7mFa2l5jfuiltuqvh2ZyZtRcwgbvbeiLO
fPIiLCgSns0j+Y2EkKWkEJzErPVnl97ZBKYrPZbdX0V9ogoL/jxEy79lo9Js29v6
RF66goJU02EJe502i7Xrs31YCKnHgvz50L6akBiadR6kkMuWvBuwYz0IZKTLqxjd
08GeRBUyalPVhtfJo3MutnaIe/ZVU7KAIwKVgom3KODcTiZYPWtV+1g/E/7p5hhv
BDG1jrIQsVkdn25fa5sd5OPkP/l0Quv5zmzPI7KS+KfeY/v4qA90m4i6vFNFTmm0
HSWWBYNTxnR1b96PITrts8MyjoXLX6Bu1VFNJPr2Jg02LVVoq6RIbe1Uo619ojA=
-----END CERTIFICATE-----

# comment between certs

-----BEGIN CERTIFICATE-----
MIIFazCCA1OgAwIBAgIUG4u+4CfiGCzPH98t08AxyVA4C5gwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMjAzMTcxMzE1MDhaFw0yMzAz
MTcxMzE1MDhaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggIiMA0GCSqGSIb3DQEB
AQUAA4ICDwAwggIKAoICAQDvZxn/+zTyW4ERvKkwWomUpi8o2tJz1dver+DJrK3g
CNlUoYzR29CWs7jK81XMsvmqL5MzTwP7Hsmd1q69FI+WcPE1aab9909IAk/GGiK2
zTleN3EQQpXn7tnyPtZaT8Y13yFHpC5RgQzTE8CR9Zi2OLEyxGQ36pA6190nxVg2
SLlafNGTZtJvN1/7r9mJhElbrUE+joeXLwNojH/nXk5W/XwzZboIHsSFVYjK2zlg
oqC6+Apo9xF9oY7nlAhQ0KKWvQVbwjGPmVN17ETodHsKJZBoXxDsZUTGCDDJCimz
W64a79lWIxiyOQ4/7Tn2FhPY0omH5UbWCPq2MnXebkOjgcvUY4NIwzpYV0W2GGGG
wwni9fllYAM9OD3bvPX5OaAWNJT6r6EassjWlv0TeGxE+BZZ+g3xPqHTGz2wHzC5
5anLDjM+4vsBVkfkV3SY7W83m74VQ+Qa3WaNXzinL0keFxupLXY8bKhEzXzSLKz+
Br8PGeGRgaSDd1kpFPg+ujN8qvgo0RDI8IqLS7zbXFoQrp1x/dWm9S8DVEXVoUA1
WQnVuWECoBS4Zf41d04pfdCttnN9zHosgwXbJ8m0TfvfkuhViuVAN/p+Mp9WnQ+H
11HEnWpSfOh7ZPjTzjuAsevVfV4g4a3kczMv1rpNPzUUPwtAqx9239wwIR9Y15cl
OQIDAQABo1MwUTAdBgNVHQ4EFgQUXIqlbj5uMUFT/jqM2kWvYZtDNkcwHwYDVR0j
BBgwFoAUXIqlbj5uMUFT/jqM2kWvYZtDNkcwDwYDVR0TAQH/BAUwAwEB/zANBgkq
hkiG9w0BAQsFAAOCAgEACM+xbb9okUGOqkQZkgtxwyQtzotXTW6yHbgfxhBwwuv+
sImOT29L9XUYnS9q8+BKAHuFkrSEIpGFUUhCUfSkle6fnGG5oNOZmCtK37thfQU9
v4Bo8qBZHjDHwk9VTtaZmAk0KbxfhugyWVCVlmDZroSCOiWGkTVhsXaKDkbw4El0
2scygbACtVxnE5Z9fSAw1OPYrYaG2InGL4/0uRez8iyuPOe5Cb/Dd49uxqsGQd3R
C7J4/oZpvoEzQRmjLboQsC90SfjhSiphGBSbaJBddl00k5VsUrqKXZSM8qQqUfV/
nlBmb2NnUlkde8KHs0PjhBho+/7fj+L7mFa2l5jfuiltuqvh2ZyZtRcwgbvbeiLO
fPIiLCgSns0j+Y2EkKWkEJzErPVnl97ZBKYrPZbdX0V9ogoL/jxEy79lo9Js29v6
RF66goJU02EJe502i7Xrs31YCKnHgvz50L6akBiadR6kkMuWvBuwYz0IZKTLqxjd
08GeRBUyalPVhtfJo3MutnaIe/ZVU7KAIwKVgom3KODcTiZYPWtV+1g/E/7p5hhv
BDG1jrIQsVkdn25fa5sd5OPkP/l0Quv5zmzPI7KS+KfeY/v4qA90m4i6vFNFTmm0
HSWWBYNTxnR1b96PITrts8MyjoXLX6Bu1VFNJPr2Jg02LVVoq6RIbe1Uo619ojA=
-----END CERTIFICATE-----

`

	// Clean certificate chain (expected normalized output)
	testCertificateChainClean = `-----BEGIN CERTIFICATE-----
MIIFazCCA1OgAwIBAgIUG4u+4CfiGCzPH98t08AxyVA4C5gwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMjAzMTcxMzE1MDhaFw0yMzAz
MTcxMzE1MDhaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggIiMA0GCSqGSIb3DQEB
AQUAA4ICDwAwggIKAoICAQDvZxn/+zTyW4ERvKkwWomUpi8o2tJz1dver+DJrK3g
CNlUoYzR29CWs7jK81XMsvmqL5MzTwP7Hsmd1q69FI+WcPE1aab9909IAk/GGiK2
zTleN3EQQpXn7tnyPtZaT8Y13yFHpC5RgQzTE8CR9Zi2OLEyxGQ36pA6190nxVg2
SLlafNGTZtJvN1/7r9mJhElbrUE+joeXLwNojH/nXk5W/XwzZboIHsSFVYjK2zlg
oqC6+Apo9xF9oY7nlAhQ0KKWvQVbwjGPmVN17ETodHsKJZBoXxDsZUTGCDDJCimz
W64a79lWIxiyOQ4/7Tn2FhPY0omH5UbWCPq2MnXebkOjgcvUY4NIwzpYV0W2GGGG
wwni9fllYAM9OD3bvPX5OaAWNJT6r6EassjWlv0TeGxE+BZZ+g3xPqHTGz2wHzC5
5anLDjM+4vsBVkfkV3SY7W83m74VQ+Qa3WaNXzinL0keFxupLXY8bKhEzXzSLKz+
Br8PGeGRgaSDd1kpFPg+ujN8qvgo0RDI8IqLS7zbXFoQrp1x/dWm9S8DVEXVoUA1
WQnVuWECoBS4Zf41d04pfdCttnN9zHosgwXbJ8m0TfvfkuhViuVAN/p+Mp9WnQ+H
11HEnWpSfOh7ZPjTzjuAsevVfV4g4a3kczMv1rpNPzUUPwtAqx9239wwIR9Y15cl
OQIDAQABo1MwUTAdBgNVHQ4EFgQUXIqlbj5uMUFT/jqM2kWvYZtDNkcwHwYDVR0j
BBgwFoAUXIqlbj5uMUFT/jqM2kWvYZtDNkcwDwYDVR0TAQH/BAUwAwEB/zANBgkq
hkiG9w0BAQsFAAOCAgEACM+xbb9okUGOqkQZkgtxwyQtzotXTW6yHbgfxhBwwuv+
sImOT29L9XUYnS9q8+BKAHuFkrSEIpGFUUhCUfSkle6fnGG5oNOZmCtK37thfQU9
v4Bo8qBZHjDHwk9VTtaZmAk0KbxfhugyWVCVlmDZroSCOiWGkTVhsXaKDkbw4El0
2scygbACtVxnE5Z9fSAw1OPYrYaG2InGL4/0uRez8iyuPOe5Cb/Dd49uxqsGQd3R
C7J4/oZpvoEzQRmjLboQsC90SfjhSiphGBSbaJBddl00k5VsUrqKXZSM8qQqUfV/
nlBmb2NnUlkde8KHs0PjhBho+/7fj+L7mFa2l5jfuiltuqvh2ZyZtRcwgbvbeiLO
fPIiLCgSns0j+Y2EkKWkEJzErPVnl97ZBKYrPZbdX0V9ogoL/jxEy79lo9Js29v6
RF66goJU02EJe502i7Xrs31YCKnHgvz50L6akBiadR6kkMuWvBuwYz0IZKTLqxjd
08GeRBUyalPVhtfJo3MutnaIe/ZVU7KAIwKVgom3KODcTiZYPWtV+1g/E/7p5hhv
BDG1jrIQsVkdn25fa5sd5OPkP/l0Quv5zmzPI7KS+KfeY/v4qA90m4i6vFNFTmm0
HSWWBYNTxnR1b96PITrts8MyjoXLX6Bu1VFNJPr2Jg02LVVoq6RIbe1Uo619ojA=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIFazCCA1OgAwIBAgIUG4u+4CfiGCzPH98t08AxyVA4C5gwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMjAzMTcxMzE1MDhaFw0yMzAz
MTcxMzE1MDhaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggIiMA0GCSqGSIb3DQEB
AQUAA4ICDwAwggIKAoICAQDvZxn/+zTyW4ERvKkwWomUpi8o2tJz1dver+DJrK3g
CNlUoYzR29CWs7jK81XMsvmqL5MzTwP7Hsmd1q69FI+WcPE1aab9909IAk/GGiK2
zTleN3EQQpXn7tnyPtZaT8Y13yFHpC5RgQzTE8CR9Zi2OLEyxGQ36pA6190nxVg2
SLlafNGTZtJvN1/7r9mJhElbrUE+joeXLwNojH/nXk5W/XwzZboIHsSFVYjK2zlg
oqC6+Apo9xF9oY7nlAhQ0KKWvQVbwjGPmVN17ETodHsKJZBoXxDsZUTGCDDJCimz
W64a79lWIxiyOQ4/7Tn2FhPY0omH5UbWCPq2MnXebkOjgcvUY4NIwzpYV0W2GGGG
wwni9fllYAM9OD3bvPX5OaAWNJT6r6EassjWlv0TeGxE+BZZ+g3xPqHTGz2wHzC5
5anLDjM+4vsBVkfkV3SY7W83m74VQ+Qa3WaNXzinL0keFxupLXY8bKhEzXzSLKz+
Br8PGeGRgaSDd1kpFPg+ujN8qvgo0RDI8IqLS7zbXFoQrp1x/dWm9S8DVEXVoUA1
WQnVuWECoBS4Zf41d04pfdCttnN9zHosgwXbJ8m0TfvfkuhViuVAN/p+Mp9WnQ+H
11HEnWpSfOh7ZPjTzjuAsevVfV4g4a3kczMv1rpNPzUUPwtAqx9239wwIR9Y15cl
OQIDAQABo1MwUTAdBgNVHQ4EFgQUXIqlbj5uMUFT/jqM2kWvYZtDNkcwHwYDVR0j
BBgwFoAUXIqlbj5uMUFT/jqM2kWvYZtDNkcwDwYDVR0TAQH/BAUwAwEB/zANBgkq
hkiG9w0BAQsFAAOCAgEACM+xbb9okUGOqkQZkgtxwyQtzotXTW6yHbgfxhBwwuv+
sImOT29L9XUYnS9q8+BKAHuFkrSEIpGFUUhCUfSkle6fnGG5oNOZmCtK37thfQU9
v4Bo8qBZHjDHwk9VTtaZmAk0KbxfhugyWVCVlmDZroSCOiWGkTVhsXaKDkbw4El0
2scygbACtVxnE5Z9fSAw1OPYrYaG2InGL4/0uRez8iyuPOe5Cb/Dd49uxqsGQd3R
C7J4/oZpvoEzQRmjLboQsC90SfjhSiphGBSbaJBddl00k5VsUrqKXZSM8qQqUfV/
nlBmb2NnUlkde8KHs0PjhBho+/7fj+L7mFa2l5jfuiltuqvh2ZyZtRcwgbvbeiLO
fPIiLCgSns0j+Y2EkKWkEJzErPVnl97ZBKYrPZbdX0V9ogoL/jxEy79lo9Js29v6
RF66goJU02EJe502i7Xrs31YCKnHgvz50L6akBiadR6kkMuWvBuwYz0IZKTLqxjd
08GeRBUyalPVhtfJo3MutnaIe/ZVU7KAIwKVgom3KODcTiZYPWtV+1g/E/7p5hhv
BDG1jrIQsVkdn25fa5sd5OPkP/l0Quv5zmzPI7KS+KfeY/v4qA90m4i6vFNFTmm0
HSWWBYNTxnR1b96PITrts8MyjoXLX6Bu1VFNJPr2Jg02LVVoq6RIbe1Uo619ojA=
-----END CERTIFICATE-----
`
)

// insertNewlines inserts newlines every n characters to simulate line-wrapped base64.
func insertNewlines(s string, n int) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && i%n == 0 {
			result.WriteRune('\n')
		}
		result.WriteRune(r)
	}
	result.WriteRune('\n') // trailing newline
	return result.String()
}

func TestNormalizeCertificate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "empty string returns empty",
			input:       "",
			expected:    "",
			expectError: false,
		},
		{
			name:        "clean certificate unchanged",
			input:       base64.StdEncoding.EncodeToString([]byte(testCertificateClean)),
			expected:    base64.StdEncoding.EncodeToString([]byte(testCertificateClean)),
			expectError: false,
		},
		{
			name:        "whitespace is stripped",
			input:       base64.StdEncoding.EncodeToString([]byte(testCertificateWithWhitespace)),
			expected:    base64.StdEncoding.EncodeToString([]byte(testCertificateClean)),
			expectError: false,
		},
		{
			name:        "comments are stripped from chain",
			input:       base64.StdEncoding.EncodeToString([]byte(testCertificateChainWithComments)),
			expected:    base64.StdEncoding.EncodeToString([]byte(testCertificateChainClean)),
			expectError: false,
		},
		{
			name:        "CRLF line endings normalized to LF",
			input:       base64.StdEncoding.EncodeToString([]byte(strings.ReplaceAll(testCertificateClean, "\n", "\r\n"))),
			expected:    base64.StdEncoding.EncodeToString([]byte(testCertificateClean)),
			expectError: false,
		},
		{
			name:        "base64 with newlines is handled",
			input:       insertNewlines(base64.StdEncoding.EncodeToString([]byte(testCertificateClean)), 76),
			expected:    base64.StdEncoding.EncodeToString([]byte(testCertificateClean)),
			expectError: false,
		},
		{
			name:        "invalid base64 returns error",
			input:       "not-valid-base64!!!",
			expected:    "",
			expectError: true,
		},
		{
			name:        "no PEM blocks returns error",
			input:       base64.StdEncoding.EncodeToString([]byte("just some text without PEM")),
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, diags := normalizeCertificate(tt.input)

			if tt.expectError {
				require.True(t, diags.HasError(), "expected error but got none")
				return
			}

			require.False(t, diags.HasError(), "unexpected error: %v", diags)
			require.Equal(t, tt.expected, result)
		})
	}
}
