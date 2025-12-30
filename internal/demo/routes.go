package demo

// RouteTemplate represents a reusable route for activities.
type RouteTemplate struct {
	Name            string
	SportTypes      []string // Which sport types can use this route
	SummaryPolyline string   // Encoded polyline (Google format)
	Distance        float64  // Approximate distance in meters
	ElevationGain   float64  // Approximate elevation gain in meters
	StartLat        float64
	StartLng        float64
	EndLat          float64
	EndLng          float64
}

// routeTemplates contains embedded route templates for demo activities.
// These are real-ish routes in the San Francisco Bay Area.
var routeTemplates = []RouteTemplate{
	// Running routes
	{
		Name:            "Golden Gate Park Loop",
		SportTypes:      []string{"Run", "Walk"},
		SummaryPolyline: "afreF`rejVmA{FgBsFkCwEuDwDeFyCgFmB_GaBaGkA_Hq@yGYcH@kHPwGh@wGdAmG~AeGtBcGjCcFdDeFxDwDjEyC|EkCnFgBbGaB`HkArH_AnIq@~I]~I?dJTrId@|HdAtHxA`HnBxGhCtFbDrE|DxDvElCfFvB`GdBtGhAtHr@lI`@|IFnJ_@~Ic@jIs@`IeA|HkApHyAxG}BjGaDtFqDfFeEtDcFdDuFrCiGlBuG~AaHpAkHh@wH@eIA",
		Distance:        8000,
		ElevationGain:   50,
		StartLat:        37.7694,
		StartLng:        -122.4862,
		EndLat:          37.7694,
		EndLng:          -122.4862,
	},
	{
		Name:            "Embarcadero Out and Back",
		SportTypes:      []string{"Run", "Walk"},
		SummaryPolyline: "k~deF~nbjVaBf@qBb@cC`@oC^yCZwCVeDR_ERuED}E?cFGqFOyF]_Gk@eGuAiGsAkFqBgF_CeFuC_EgDiDuDsCgEaCuEoBaFmAgF{@cGg@eGIeGB_GZ}Fn@{FjA}EjByE|BeEvCaDfDoC~D}AlEsAlFy@bFc@fFCfFh@bF~@|EhBzEtBvEfCnEzCtDnDjDfE~CzEdC`FhCjF`CrFtBnFpBhFfBxFn@zFN|FKxFo@tFaArEuApE_BfDgC~CqCfC}CjBiDb@wD",
		Distance:        10000,
		ElevationGain:   20,
		StartLat:        37.7936,
		StartLng:        -122.3930,
		EndLat:          37.8044,
		EndLng:          -122.4220,
	},
	{
		Name:            "Presidio Hills",
		SportTypes:      []string{"Run", "TrailRun", "Hike"},
		SummaryPolyline: "gyeeF~odjVkCiByBuB}AoCqAgDcA{Dq@eFa@gFEiFZeFp@cFdA}ErAwEdBmEtBaE~BqDlCaDbD{CdDkCtDyBdEgBnEu@zEe@bFCfFVdF|@bFfAfF`BdEzBzDhCvDxCnD`DfDfDvChD~BjDlBpDdAtDj@vDDtDa@rD}@pDeAfDkBdDwBvC_DlC}C~B{DjBoE~AsEnAmF`AoFn@qF\\sFBsF",
		Distance:        7500,
		ElevationGain:   200,
		StartLat:        37.7989,
		StartLng:        -122.4662,
		EndLat:          37.7899,
		EndLng:          -122.4774,
	},
	{
		Name:            "Marina Green Short",
		SportTypes:      []string{"Run", "Walk"},
		SummaryPolyline: "cweeF|xdjV_@lBYpB]nBa@lBa@hBg@dBo@~As@xAy@rAaAlA_A`AaBz@cBj@eBb@gBVgBHiBAgBQeBa@_Bk@yAs@sAaAmAeAaAiAu@mAs@qAi@wAa@{AY_BO_BG_B@_BLaBZcBh@gBr@gBz@eBfAaBjAyAtAmAvAkA`BaAbBu@jBm@pBa@rBWtBO",
		Distance:        3000,
		ElevationGain:   10,
		StartLat:        37.8070,
		StartLng:        -122.4361,
		EndLat:          37.8070,
		EndLng:          -122.4361,
	},

	// Cycling routes
	{
		Name:            "Hawk Hill Loop",
		SportTypes:      []string{"Ride", "GravelRide"},
		SummaryPolyline: "ssfeF`ifjVuCjA}ClAeDbAgDr@mDd@uDTyD@{DEwDUsDe@oDw@mDaAiDgAcDoA}CyAuC_BmCgBeC}BqBoCiBkC{AmCkAoCaA{Cq@aDe@eDQ_E?}DR{Dj@wDdAqDrA}CdBoC|BiCfCoBpCcB~CqApDaAdEo@xE]hFIzE?lEDfE^dEp@jErA`EdBzDfCjDzCvClDjCzDfCbE~BhEvBnExAtE|@~E`@xEBpEe@lEiApDcBzCyBnC{C|BgDjBwDpAaDz@aDh@mDV{DJmEEoEa@oEcAmEuAgE}AcE}B{DqC",
		Distance:        25000,
		ElevationGain:   400,
		StartLat:        37.8324,
		StartLng:        -122.4795,
		EndLat:          37.8324,
		EndLng:          -122.4795,
	},
	{
		Name:            "Paradise Loop",
		SportTypes:      []string{"Ride"},
		SummaryPolyline: "yngeF|dgjVmBuAaBgBiAoBw@wB_@aCCoCTeCn@eCdAeCrAeCfB}BzBkBfCaB|CkApDaAlE{@bF{@vFs@`Go@`Ha@~Ga@dHU~GGxGBtGXpGr@fGhAlFzAhFdBlEvBpD~BjDlC`DzCtCjDbCrDjBzDfBbE`BdEtAdEfAnEv@pEh@pE\\nELjECjEUjEi@hEu@fEeAbE_BbEaBhE_BpEwA~EoAbFkAbFaAtFq@vFe@vFQtF?rFLpFd@lFx@jFnAfF~AlE`BxDzBzD`CvDjClDrCfDxC~CbDrCfDjChDdChDzBjD`CnDhC~DpCrEtC~EtCfFnClFhChFnB`FxApE|@fEd@rEFlEC|Ea@nEu@fEcA~DiAtD{AbDgBzCwB|CsB~CoB~CoCdDoC~CoCvCqClCsCxBoC~AyCfAaDbAsDn@{D^_EBgEYkEq@kEcAmEoAiEaB_EoByD}BeD}BaDgCaDeC_DsC_D_DcD_D{CgD{CsD{CcE{CsE{C_F{CeFgD",
		Distance:        50000,
		ElevationGain:   500,
		StartLat:        37.8599,
		StartLng:        -122.4867,
		EndLat:          37.8599,
		EndLng:          -122.4867,
	},
	{
		Name:            "Mt Tam Climb",
		SportTypes:      []string{"Ride", "GravelRide", "MountainBikeRide"},
		SummaryPolyline: "cnfeF~yhjVuDdBaDvBsCfC_CrCcBlD{AfE}@bFi@`GQ~FBhGXrGr@nGfAfGvArFfBvE~BjEdCzD`DhDjDzC~DrClE~BvEhBnF|AbGhArGt@|Gb@bH@~Ge@`HaAhH_BvGeBlGeC~FgCfFoDnE_E~DcEfDgEvCuEjCaFvBcFxAgFhAkFt@wFd@{FL{FEyF_@wFu@qF_AkFmAeFaBaF{ByE_CmEcCcEmC{DwCoD_D{CgDkCqD_CaE_CqEgC_FkCqFgC_GoByGmAeHw@mH]qH?sHZsHv@oHjA",
		Distance:        35000,
		ElevationGain:   800,
		StartLat:        37.9023,
		StartLng:        -122.5270,
		EndLat:          37.9235,
		EndLng:          -122.5957,
	},
	{
		Name:            "Flat Bay Trail",
		SportTypes:      []string{"Ride", "VirtualRide"},
		SummaryPolyline: "eydeF~tbjVmBXsBNuB@uBKsBYqBg@mBu@iBaAeBoAaBwA}A_BwAkBmAsBoAyBiA}BcAcC}@gCu@kCm@oCe@sCY{CQaDGgD?kDBmDJoDRoD\\oD`@kDj@iDr@eD|@aDfA}CpAwCzAqCdBkCnBcCxByBbCmBlCiBvC_BbDqAfDiAjD_ArDs@zDk@~Da@dEUdEKbE?bELbEXbE`@`EhAzDfB~CfCtCzCbCnDhBnDnA|DrA~DpAtE~@~Ej@jF\\vFFxF",
		Distance:        20000,
		ElevationGain:   30,
		StartLat:        37.7746,
		StartLng:        -122.3892,
		EndLat:          37.7746,
		EndLng:          -122.3892,
	},

	// Hike/Trail routes
	{
		Name:            "Lands End Trail",
		SportTypes:      []string{"Hike", "TrailRun", "Walk"},
		SummaryPolyline: "gvdeFxtfjV~@fBl@hBZ`BHrBAhBWdBk@`B}@tAkA~@gApA_AbBu@zBe@fCQ~B?nBXrBt@jBbA~AlAbAzAl@jBTjB?lB[jBu@`BmAhAyA|@eBh@qBXwBD}BM}B_@}Bm@wBaA}AaAcBsAqBqAqByAmBeBgBkBaBqB}AwBiA{B{@sCi@mCUqC?qCRmCf@oCdA}BfAmBdBkBrBgBpBgBdBgB|AaBnAsAzBmAnC_AxC",
		Distance:        6000,
		ElevationGain:   150,
		StartLat:        37.7780,
		StartLng:        -122.5110,
		EndLat:          37.7852,
		EndLng:          -122.5050,
	},

	// Swim (simple point-to-point in open water)
	{
		Name:            "Aquatic Park Swim",
		SportTypes:      []string{"Swim"},
		SummaryPolyline: "wkfeFjfejVeCx@qCZ}C?qCa@eCy@qCsAaC_B}BiBoBeC",
		Distance:        1500,
		ElevationGain:   0,
		StartLat:        37.8070,
		StartLng:        -122.4229,
		EndLat:          37.8084,
		EndLng:          -122.4195,
	},
}
