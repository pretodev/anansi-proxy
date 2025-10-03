" Vim syntax file
" Language: ApiMock
" Maintainer: Anansi Proxy
" Latest Revision: 2025-10-03

if exists("b:current_syntax")
  finish
endif

" Comments
syntax match apimockComment "#.*$"

" HTTP Methods
syntax keyword apimockMethod GET POST PUT DELETE PATCH HEAD OPTIONS TRACE CONNECT
syntax match apimockMethodLine "^\s*\(GET\|POST\|PUT\|DELETE\|PATCH\|HEAD\|OPTIONS\|TRACE\|CONNECT\)\s\+"

" Response line: -- 200: Description
syntax match apimockResponseSeparator "^\s*--"
syntax match apimockStatusCode "\(^\s*--\s*\)\@<=\d\{3\}"
syntax match apimockResponseLine "^\s*--\s*\d\{3\}\s*:\s*.*$" contains=apimockResponseSeparator,apimockStatusCode,apimockDescription
syntax match apimockDescription "\(^\s*--\s*\d\{3\}\s*:\s*\)\@<=.*$" contained

" Path and path parameters
syntax match apimockPath "\/[^\s]*" contains=apimockPathParam
syntax match apimockPathParam "{\w\+}" contained

" Query parameters
syntax match apimockQueryParam "?\w\+=\w\+"
syntax match apimockQueryParam "&\w\+=\w\+"

" Properties (headers)
syntax match apimockProperty "^\s*[A-Za-z][A-Za-z0-9_\-\.]*\s*:" contains=apimockPropertyKey
syntax match apimockPropertyKey "^\s*[A-Za-z][A-Za-z0-9_\-\.]*" contained
syntax match apimockPropertyValue ":\s*\zs.*$"

" Strings
syntax region apimockString start='"' end='"' skip='\\"'
syntax region apimockString start="'" end="'" skip="\\'"

" Numbers
syntax match apimockNumber "\<\d\+\>"
syntax match apimockNumber "\<\d\+\.\d\+\>"

" Booleans
syntax keyword apimockBoolean true false null

" JSON body detection (simple)
syntax region apimockJSON start="^\s*{" end="^\s*}" transparent contains=apimockJSONKey,apimockString,apimockNumber,apimockBoolean
syntax match apimockJSONKey '"\w\+"\s*:' contained

" XML body detection
syntax match apimockXMLTag "<[^>]\+>" contained
syntax region apimockXML start="^\s*<?\?xml" end="^\s*<\/[^>]\+>\s*$" transparent contains=apimockXMLTag

" Operators and special characters
syntax match apimockOperator "[{}[\]:,]"

" Define highlighting
highlight default link apimockComment Comment
highlight default link apimockMethod Keyword
highlight default link apimockMethodLine Keyword
highlight default link apimockResponseSeparator Keyword
highlight default link apimockStatusCode Number
highlight default link apimockDescription String
highlight default link apimockPath String
highlight default link apimockPathParam Identifier
highlight default link apimockQueryParam Identifier
highlight default link apimockPropertyKey Type
highlight default link apimockPropertyValue String
highlight default link apimockProperty Type
highlight default link apimockString String
highlight default link apimockNumber Number
highlight default link apimockBoolean Boolean
highlight default link apimockJSONKey Type
highlight default link apimockXMLTag Tag
highlight default link apimockOperator Operator

let b:current_syntax = "apimock"
