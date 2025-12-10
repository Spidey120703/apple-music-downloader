package cdm

import "encoding/base64"

const (
	DefaultPrivateKey = "-----BEGIN RSA PRIVATE KEY-----\n" +
		"MIIEpAIBAAKCAQEA2bO3yvFwNnIHsbDl3MTjKdDsiBWsuZWOGVxInFWAVMp+nffG\n" +
		"YlquTKpJurEry95yprcRB3hYhvA5ghsACidcWPDEPVqqRZ7YXLevyUA+Sn2Jxpvt\n" +
		"OcwyFHbSwruNxprWOkHCT774O4L/wJUt5x2C4iFCrJByjw0omN8u+EHdavvH7ZPn\n" +
		"b3/EZp/cpZa9/+HOkutvBHBvaPp18F8JQhzUQ9MwLuDFTr+QLDB5+Y57Je2tNYDK\n" +
		"xD1K+Ed5Ja0A4OKhPKIwPwPre0nt5scjLba3LSAKtKxiGqFtWO4U7Tf1YrdjJv2o\n" +
		"9o8Sf8qcnbpzvQ4KwFqehuJnB7+W7mdJJw12PQIDAQABAoIBACE32wOMc6LbI3Fp\n" +
		"nKljIYZv6qeZJxHqUBRukGXKZhqKC2fvNsYrMA1irn1eK2CgQL5PkLmjE18DqMLB\n" +
		"e/AQsXagxlDWVMTqx/jdzmTW+KpFHZDAmiIHllypBN/R3oA/gBDDl/KzIQ1zn7Kz\n" +
		"EJ4DUsVObe4G3HQXfepVo8Udx7tbB7X6wHe2kEgFyY3lPdvubik0C4t4ipSD79y7\n" +
		"SfW7XVA5XUQmqN4U2kWM0uSwzd4BA7hqyScJsygf6KgpMWPS2xFZEZQRUpYcBH48\n" +
		"E7YqNrrlYP3yaQ+9Jx56kKS0mvv3vUXS7AfUbU8CiHwD9I3BGwswEUueOGGVeXbx\n" +
		"tFF8s8ECgYEA97BDcL/bt+r3qJF0dxtMB5ZngJbFx9RdsblYepVpblr2UfxnFttO\n" +
		"PoNSKa4W36HuDsun49dkaoABJWdtZs2Hy6q+xvEgozvhMaBVE3spnWnzCT1yTMYL\n" +
		"G02uDEl0dPiTg116bVElaswtqMXvnnpbOTMTe7Ig9sWiUW/GH9RM+N8CgYEA4QHb\n" +
		"+OA0BfczbVQP9B+plt4mAuu4BDm4GPwq1yXOWo3Ct8Ik+HeY1hqOObpfyQMAza+E\n" +
		"e/kP6W8vXpiElGrmiUbTXK4Rzmf+yYeOrvl3D80bFq4GtDNAIQD3jpj6zjlT+Gzw\n" +
		"I501gRx5iPl4fSccRSdpoeri7F9ANtc6EEGFyGMCgYEAjMznWYXHGkL47BtbkIW0\n" +
		"769BQSj0X4dKh8gsEusylugglDSeSbD7RrASGd175T7A/CorU2rTC3OesyubVlBJ\n" +
		"/K4gaykRe5mDh1l0Y3GlE3XyEXObsSb3k1rSMOvkxsWz3X5bJR923MIaxpFWiMlX\n" +
		"aCmvzqZQ9NceUZrvjpJ5+xMCgYAJa8KCESEcftUwZqykVA8Nug9tX+E8jA4hPa2t\n" +
		"hG+3augUOZTCsn87t7Dsydjo2a9W7Vpmtm7sHzOkik5CyJcOeGCxKLimI8SPO5XF\n" +
		"zbwmdTgFIxQ0x1CQETJMTityJwRVCnqjgxmSZlbQXWGmG9UbMCNEHEmUDAjsQuaz\n" +
		"d4racQKBgQDR1Y2kalvleYGrhwcA8LTnIh0rYEfAt9YxNmTi5qDKf5QPvUP2v+WO\n" +
		"fSB5coUqR8LBweHE5V8JgFt74fdLBqZV/k2z/dI0r+EQWmpZ2uPEC0Khk/Sb9iRD\n" +
		"fH7at3PMusrkwZCGZ8beFEAr6icXclV08nPCNGB6WckacfzpAj8Azg==\n" +
		"-----END RSA PRIVATE KEY-----"

	DefaultClientIDBase64 = "" +
		"CAESmgsK3QMIAhIQeeRrycR5oAnVvSCrdzFrTxivgsKlBiKOAjCCAQoCggEBANmz" +
		"t8rxcDZyB7Gw5dzE4ynQ7IgVrLmVjhlcSJxVgFTKfp33xmJarkyqSbqxK8vecqa3" +
		"EQd4WIbwOYIbAAonXFjwxD1aqkWe2Fy3r8lAPkp9icab7TnMMhR20sK7jcaa1jpB" +
		"wk+++DuC/8CVLecdguIhQqyQco8NKJjfLvhB3Wr7x+2T529/xGaf3KWWvf/hzpLr" +
		"bwRwb2j6dfBfCUIc1EPTMC7gxU6/kCwwefmOeyXtrTWAysQ9SvhHeSWtAODioTyi" +
		"MD8D63tJ7ebHIy22ty0gCrSsYhqhbVjuFO039WK3Yyb9qPaPEn/KnJ26c70OCsBa" +
		"nobiZwe/lu5nSScNdj0CAwEAASjwIkgBUqoBCAEQABqBAQQZhh0LPs5wmuuobaJo" +
		"fVK1k0DjvnNhqvOMfGw0Zlzum4aTAvasMiyWfhjo/+xmHtsRvK3ek9EOdIB1e2c5" +
		"azFuScAMS2n7ZGzqA8XBb+UPM46FUeGt7o1jDm/AysaZt4U6Ji8wXl41dWA9kF/i" +
		"IK7uThSmb+mhspLLYo3AUiu2hiIgFm8idU4+UvSfVB4JveJ+hqeNbpYuNWkrxlbj" +
		"9DDjWgYSgAIemDQcy+RKUwwGq59NhaxYSH3hxSHGCkhcXnjNC0OeV5gBdJQl7uqN" +
		"90lkF3JxnlvYF3mhux7pZR5jii4KaNG6+vZXEq21irNMnoSxwIlzvpMov7xOvQWV" +
		"m00K+xDkO20ncTC1ClXpmAAHyDXmMeTrzvCLo7tc3USbaImlIWAX92saZojzJ3n9" +
		"gc+cjBKGqz2AgcsFCigSZ5vpLtz/wEk5PxIGKJ6OWjEy4D5HZG0p2MYyhM84fUh3" +
		"TOfuexK1ceWrOfPxCbxSPRi9w0BEaDmixt/K4mIalUFTBJsWxtE6ww38UmFLktWo" +
		"MM8+QLnhxe6jmuVpuchdLtnMPnkAs6XjGrQFCq4CCAESEGnj6Ji7LD+4o7MoHYT4" +
		"jBQYjtW+kQUijgIwggEKAoIBAQDY9um1ifBRIOmkPtDZTqH+CZUBbb0eK0Cn3NHF" +
		"f8MFUDzPEz+emK/OTub/hNxCJCao//pP5L8tRNUPFDrrvCBMo7Rn+iUb+mA/2yXi" +
		"J6ivqcN9Cu9i5qOU1ygon9SWZRsujFFB8nxVreY5Lzeq0283zn1Cg1stcX4tOHT7" +
		"utPzFG/ReDFQt0O/GLlzVwB0d1sn3SKMO4XLjhZdncrtF9jljpg7xjMIlnWJUqxD" +
		"o7TQkTytJmUl0kcM7bndBLerAdJFGaXc6oSY4eNy/IGDluLCQR3KZEQsy/mLeV1g" +
		"gQ44MFr7XOM+rd+4/314q/deQbjHqjWFuVr8iIaKbq+R63ShAgMBAAEo8CISgAMi" +
		"i2Mw6z+Qs1bvvxGStie9tpcgoO2uAt5Zvv0CDXvrFlwnSbo+qR71Ru2IlZWVSbN5" +
		"XYSIDwcwBzHjY8rNr3fgsXtSJty425djNQtF5+J2jrAhf3Q2m7EI5aohZGpD2E0c" +
		"r+dVj9o8x0uJR2NWR8FVoVQSXZpad3M/4QzBLNto/tz+UKyZwa7Sc/eTQc2+ZcDS" +
		"3ZEO3lGRsH864Kf/cEGvJRBBqcpJXKfG+ItqEW1AAPptjuggzmZEzRq5xTGf6or+" +
		"bXrKjCpBS9G1SOyvCNF1k5z6lG8KsXhgQxL6ADHMoulxvUIihyPY5MpimdXfUdEQ" +
		"5HA2EqNiNVNIO4qP007jW51yAeThOry4J22xs8RdkIClOGAauLIl0lLA4flMzW+V" +
		"fQl5xYxP0E5tuhn0h+844DslU8ZF7U1dU2QprIApffXD9wgAACk26Rggy8e96z8i" +
		"86/+YYyZQkc9hIdCAERrgEYCEbByzONrdRDs1MrS/ch1moV5pJv63BIKvQHGvLka" +
		"FgoMY29tcGFueV9uYW1lEgZHb29nbGUaIQoKbW9kZWxfbmFtZRITQU9TUCBvbiBJ" +
		"QSBFbXVsYXRvchoYChFhcmNoaXRlY3R1cmVfbmFtZRIDeDg2Gh4KC2RldmljZV9u" +
		"YW1lEg9nZW5lcmljX3g4Nl9hcm0aIgoMcHJvZHVjdF9uYW1lEhJzZGtfZ3Bob25l" +
		"X3g4Nl9hcm0aZAoKYnVpbGRfaW5mbxJWZ29vZ2xlL3Nka19ncGhvbmVfeDg2X2Fy" +
		"bS9nZW5lcmljX3g4Nl9hcm06OS9QU1IxLjE4MDcyMC4xMjIvNjczNjc0Mjp1c2Vy" +
		"ZGVidWcvZGV2LWtleXMaHgoUd2lkZXZpbmVfY2RtX3ZlcnNpb24SBjE0LjAuMBok" +
		"Ch9vZW1fY3J5cHRvX3NlY3VyaXR5X3BhdGNoX2xldmVsEgEwMg4QASAAKA0wAEAA" +
		"SABQAA=="
)

var DefaultClientID []byte

func init() {
	var err error
	DefaultClientID, err = base64.StdEncoding.DecodeString(DefaultClientIDBase64)
	if err != nil {
		panic(err)
	}
}
