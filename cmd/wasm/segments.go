//go:build js && wasm

package main

// ============================================================================
// Segment Countries (no generated version - uses ListCountries directly)
// ============================================================================

//wasm:category Segment Efforts

// getSegmentCountries returns country statistics for segments
// Called from JS: goStorage.getSegmentCountries()
//
//wasm:export
var getSegmentCountries = wrapWasmAthlete("getSegmentCountries", func(wc *WasmContext) interface{} {
	result, err := wc.Registry.SegmentsService.ListCountries(wc.Ctx, wc.AthleteID)
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(result)
})
