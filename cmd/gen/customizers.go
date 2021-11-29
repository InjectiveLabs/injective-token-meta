package main

type tokenMetaCustomizer func(meta *Token)

var tokenMetaCustomizerMap = map[string]tokenMetaCustomizer{}
