package stoken

import "errors"

var ErrTokenInvalid = errors.New("stoken: token invalid")
var ErrTokenInvalidClaims = errors.New("stoken: token invalid claims")
var ErrParsingData = errors.New("stoken: parsing data")
