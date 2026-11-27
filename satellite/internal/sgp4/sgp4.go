package sgp4

import "math"

// Elements 内核历元与平均根数，角度为弧度 / kernel epoch and mean elements in radians.
type Elements struct {
	EpochDays, BStar, Eccentricity                                         float64
	InclinationRad, NodeRad, ArgumentRad, MeanAnomalyRad                   float64
	MeanMotionRadPerMin, MeanMotionDotRadPerMin2, MeanMotionDDotRadPerMin3 float64
}

// Record 初始化后的传播系数 / initialized propagation coefficients.
type Record struct {
	propagationState
	a               float64
	alta            float64
	altp            float64
	argpdot         float64
	argpoRad        float64
	bstar           float64
	cc1             float64
	cc4             float64
	cc5             float64
	d2              float64
	d2201           float64
	d2211           float64
	d3              float64
	d3210           float64
	d3222           float64
	d4              float64
	d4410           float64
	d4422           float64
	d5220           float64
	d5232           float64
	d5421           float64
	d5433           float64
	dedt            float64
	del1            float64
	del2            float64
	del3            float64
	delmo           float64
	didt            float64
	dmdt            float64
	dnodt           float64
	domdt           float64
	e3              float64
	ecco            float64
	ee2             float64
	eta             float64
	gstoRad         float64
	incloRad        float64
	init            byte
	irez            int
	isimp           int
	j2              float64
	j3              float64
	j3oj2           float64
	j4              float64
	mdot            float64
	method          byte
	moRad           float64
	mus             float64
	nddot           float64
	ndot            float64
	meanMotionKozai float64
	meanMotion      float64
	nodecf          float64
	nodedot         float64
	nodeoRad        float64
	omgcof          float64
	mode            byte
	peo             float64
	pgho            float64
	pho             float64
	pinco           float64
	plo             float64
	earthRadiusKM   float64
	se2             float64
	se3             float64
	sgh2            float64
	sgh3            float64
	sgh4            float64
	sh2             float64
	sh3             float64
	si2             float64
	si3             float64
	sinmao          float64
	sl2             float64
	sl3             float64
	sl4             float64
	t2cof           float64
	t3cof           float64
	t4cof           float64
	t5cof           float64
	tumin           float64
	xfact           float64
	xgh2            float64
	xgh3            float64
	xgh4            float64
	xh2             float64
	xh3             float64
	xi2             float64
	xi3             float64
	xke             float64
	xl2             float64
	xl3             float64
	xl4             float64
	xlamoRad        float64
	xmcof           float64
	zmolRad         float64
	zmosRad         float64
}

type propagationState struct {
	t, am, em, inclinationMeanRad, nodeMeanRad, om, mmRad, nm float64
	con41, x1mth2, x7thm1, aycof, xlcof, atime, xliRad, xni   float64
	error                                                     int
}

// Init 重置并初始化固定引力模型的传播系数 / reset and initialize fixed-gravity propagation coefficients.
func (s *Record) Init(e Elements) {
	*s = Record{}
	initialize(1, 'i', "", e.EpochDays, e.BStar, e.MeanMotionDotRadPerMin2, e.MeanMotionDDotRadPerMin3, e.Eccentricity, e.ArgumentRad, e.InclinationRad, e.MeanAnomalyRad, e.MeanMotionRadPerMin, e.NodeRad, s)
}

// Propagate 按距历元的分钟数求TEME状态与错误码 / TEME state and status at minutes since epoch.
func (s *Record) Propagate(minutes float64) (r, v [3]float64, code int) {
	work := s.propagationState
	propagate(s, &work, minutes, &r, &v)
	return r, v, work.error
}

func gravity(_ int, tumin, mu, radiusKM, xke, j2, j3, j4, j3oj2 *float64) {
	*mu = 398600.8
	*radiusKM = 6378.135
	*xke = 60.0 / math.Sqrt(*radiusKM**radiusKM**radiusKM / *mu)
	*tumin = 1.0 / *xke
	*j2 = 0.001082616
	*j3 = -0.00000253881
	*j4 = -0.00000165597
	*j3oj2 = *j3 / *j2
}

// GSTime UT1儒略日的格林尼治平恒星时，弧度 / Greenwich mean sidereal time in radians at UT1 Julian date.
func GSTime(jdut1 float64) float64 {
	var twopi, deg2rad, temp, tut1 float64
	twopi = 2.0 * 3.14159265358979323846
	deg2rad = 3.14159265358979323846 / 180.0
	tut1 = (jdut1 - 2451545.0) / 36525.0
	temp = -6.2e-6*tut1*tut1*tut1 + 0.093104*tut1*tut1 + (876600.0*3600+8640184.812866)*tut1 + 67310.54841
	temp = math.Mod(temp*deg2rad/240.0, twopi)
	if temp < 0.0 {
		temp += twopi
	}
	return temp
}

func initl(xke float64, j2 float64, ecco float64, epoch float64, incloRad float64, meanMotionKozai float64, opsmode byte, method *byte, ainv *float64, ao *float64, con41 *float64, con42 *float64, cosio *float64, cosio2 *float64, eccsq *float64, omeosq *float64, posq *float64, rp *float64, rteosq *float64, sinio *float64, gstoRad *float64, meanMotion *float64) {
	var ak, d1, del, adel, po, x2o3, ds70, ts70, tfrac, c1, thgr70, fk5r, c1p2p, twopi, gsto1Rad float64
	twopi = 2.0 * 3.14159265358979323846
	x2o3 = 2.0 / 3.0
	*eccsq = ecco * ecco
	*omeosq = 1.0 - *eccsq
	*rteosq = math.Sqrt(*omeosq)
	*cosio = math.Cos(incloRad)
	*cosio2 = *cosio * *cosio
	ak = math.Pow(xke/meanMotionKozai, x2o3)
	d1 = 0.75 * j2 * (3.0**cosio2 - 1.0) / (*rteosq * *omeosq)
	del = d1 / (ak * ak)
	adel = ak * (1.0 - del*del - del*(1.0/3.0+134.0*del*del/81.0))
	del = d1 / (adel * adel)
	*meanMotion = meanMotionKozai / (1.0 + del)
	*ao = math.Pow(xke / *meanMotion, x2o3)
	*sinio = math.Sin(incloRad)
	po = *ao * *omeosq
	*con42 = 1.0 - 5.0**cosio2
	*con41 = -*con42 - *cosio2 - *cosio2
	*ainv = 1.0 / *ao
	*posq = po * po
	*rp = *ao * (1.0 - ecco)
	*method = 'n'
	ts70 = epoch - 7305.0
	ds70 = math.Floor(ts70 + 1.0e-8)
	tfrac = ts70 - ds70
	c1 = 1.72027916940703639e-2
	thgr70 = 1.7321343856509374
	fk5r = 5.07551419432269442e-15
	c1p2p = c1 + twopi
	gsto1Rad = math.Mod(thgr70+c1*ds70+c1p2p*tfrac+ts70*ts70*fk5r, twopi)
	if gsto1Rad < 0.0 {
		gsto1Rad = gsto1Rad + twopi
	}
	*gstoRad = GSTime(epoch + 2433281.5)
}

func initialize(whichconst int, opsmode byte, satn string, epoch float64, xbstar float64, xndot float64, xnddot float64, xecco float64, xargpo float64, xinclo float64, xmo float64, xno_kozai float64, xnodeo float64, satrec *Record) bool {
	var r, v [3]float64
	var ao, ainv, con42, cosio, sinio, cosio2, eccsq, omeosq, posq, rp, rteosq, cnodm, snodm, cosim, sinim, cosomm, sinomm, cc1sq, cc2, cc3, coef, coef1, cosio4, day, dndt, em, emsq, eeta, etasq, gamRad, argpmRad, nodemRad, inclmRad, mmRad, nm, perige, pinvsq, psisq, qzms24, rtemsq, s1, s2, s3, s4, s5, s6, s7, sfour, ss1, ss2, ss3, ss4, ss5, ss6, ss7, sz1, sz2, sz3, sz11, sz12, sz13, sz21, sz22, sz23, sz31, sz32, sz33, tc, temp, temp1, temp2, temp3, tsi, xpidot, xhdot1, z1, z2, z3, z11, z12, z13, z21, z22, z23, z31, z32, z33, qzms2t, ss, x2o3, delmotemp, qzms2ttemp, qzms24temp, temp4 float64
	temp4 = 1.5e-12
	satrec.isimp = 0
	satrec.method = 'n'
	satrec.aycof = 0.0
	satrec.con41 = 0.0
	satrec.cc1 = 0.0
	satrec.cc4 = 0.0
	satrec.cc5 = 0.0
	satrec.d2 = 0.0
	satrec.d3 = 0.0
	satrec.d4 = 0.0
	satrec.delmo = 0.0
	satrec.eta = 0.0
	satrec.argpdot = 0.0
	satrec.omgcof = 0.0
	satrec.sinmao = 0.0
	satrec.t = 0.0
	satrec.t2cof = 0.0
	satrec.t3cof = 0.0
	satrec.t4cof = 0.0
	satrec.t5cof = 0.0
	satrec.x1mth2 = 0.0
	satrec.x7thm1 = 0.0
	satrec.mdot = 0.0
	satrec.nodedot = 0.0
	satrec.xlcof = 0.0
	satrec.xmcof = 0.0
	satrec.nodecf = 0.0
	satrec.irez = 0
	satrec.d2201 = 0.0
	satrec.d2211 = 0.0
	satrec.d3210 = 0.0
	satrec.d3222 = 0.0
	satrec.d4410 = 0.0
	satrec.d4422 = 0.0
	satrec.d5220 = 0.0
	satrec.d5232 = 0.0
	satrec.d5421 = 0.0
	satrec.d5433 = 0.0
	satrec.dedt = 0.0
	satrec.del1 = 0.0
	satrec.del2 = 0.0
	satrec.del3 = 0.0
	satrec.didt = 0.0
	satrec.dmdt = 0.0
	satrec.dnodt = 0.0
	satrec.domdt = 0.0
	satrec.e3 = 0.0
	satrec.ee2 = 0.0
	satrec.peo = 0.0
	satrec.pgho = 0.0
	satrec.pho = 0.0
	satrec.pinco = 0.0
	satrec.plo = 0.0
	satrec.se2 = 0.0
	satrec.se3 = 0.0
	satrec.sgh2 = 0.0
	satrec.sgh3 = 0.0
	satrec.sgh4 = 0.0
	satrec.sh2 = 0.0
	satrec.sh3 = 0.0
	satrec.si2 = 0.0
	satrec.si3 = 0.0
	satrec.sl2 = 0.0
	satrec.sl3 = 0.0
	satrec.sl4 = 0.0
	satrec.gstoRad = 0.0
	satrec.xfact = 0.0
	satrec.xgh2 = 0.0
	satrec.xgh3 = 0.0
	satrec.xgh4 = 0.0
	satrec.xh2 = 0.0
	satrec.xh3 = 0.0
	satrec.xi2 = 0.0
	satrec.xi3 = 0.0
	satrec.xl2 = 0.0
	satrec.xl3 = 0.0
	satrec.xl4 = 0.0
	satrec.xlamoRad = 0.0
	satrec.zmolRad = 0.0
	satrec.zmosRad = 0.0
	satrec.atime = 0.0
	satrec.xliRad = 0.0
	satrec.xni = 0.0
	gravity(whichconst, &satrec.tumin, &satrec.mus, &satrec.earthRadiusKM, &satrec.xke, &satrec.j2, &satrec.j3, &satrec.j4, &satrec.j3oj2)
	satrec.error = 0
	satrec.mode = opsmode
	satrec.bstar = xbstar
	satrec.ndot = xndot
	satrec.nddot = xnddot
	satrec.ecco = xecco
	satrec.argpoRad = xargpo
	satrec.incloRad = xinclo
	satrec.moRad = xmo
	satrec.meanMotionKozai = xno_kozai
	satrec.nodeoRad = xnodeo
	satrec.nm = 0.0
	satrec.mmRad = satrec.nm
	satrec.nodeMeanRad = satrec.mmRad
	satrec.inclinationMeanRad = satrec.nodeMeanRad
	satrec.em = satrec.inclinationMeanRad
	satrec.am = satrec.em
	ss = 78.0/satrec.earthRadiusKM + 1.0
	qzms2ttemp = (120.0 - 78.0) / satrec.earthRadiusKM
	qzms2t = qzms2ttemp * qzms2ttemp * qzms2ttemp * qzms2ttemp
	x2o3 = 2.0 / 3.0
	satrec.init = 'y'
	satrec.t = 0.0
	initl(satrec.xke, satrec.j2, satrec.ecco, epoch, satrec.incloRad, satrec.meanMotionKozai, satrec.mode, &satrec.method, &ainv, &ao, &satrec.con41, &con42, &cosio, &cosio2, &eccsq, &omeosq, &posq, &rp, &rteosq, &sinio, &satrec.gstoRad, &satrec.meanMotion)
	satrec.a = math.Pow(satrec.meanMotion*satrec.tumin, -2.0/3.0)
	satrec.alta = satrec.a*(1.0+satrec.ecco) - 1.0
	satrec.altp = satrec.a*(1.0-satrec.ecco) - 1.0
	satrec.error = 0
	if omeosq >= 0.0 || satrec.meanMotion >= 0.0 {
		satrec.isimp = 0
		if rp < 220.0/satrec.earthRadiusKM+1.0 {
			satrec.isimp = 1
		}
		sfour = ss
		qzms24 = qzms2t
		perige = (rp - 1.0) * satrec.earthRadiusKM
		if perige < 156.0 {
			sfour = perige - 78.0
			if perige < 98.0 {
				sfour = 20.0
			}
			qzms24temp = (120.0 - sfour) / satrec.earthRadiusKM
			qzms24 = qzms24temp * qzms24temp * qzms24temp * qzms24temp
			sfour = sfour/satrec.earthRadiusKM + 1.0
		}
		pinvsq = 1.0 / posq
		tsi = 1.0 / (ao - sfour)
		satrec.eta = ao * satrec.ecco * tsi
		etasq = satrec.eta * satrec.eta
		eeta = satrec.ecco * satrec.eta
		psisq = math.Abs(1.0 - etasq)
		coef = qzms24 * math.Pow(tsi, 4.0)
		coef1 = coef / math.Pow(psisq, 3.5)
		cc2 = coef1 * satrec.meanMotion * (ao*(1.0+1.5*etasq+eeta*(4.0+etasq)) + 0.375*satrec.j2*tsi/psisq*satrec.con41*(8.0+3.0*etasq*(8.0+etasq)))
		satrec.cc1 = satrec.bstar * cc2
		cc3 = 0.0
		if satrec.ecco > 1.0e-4 {
			cc3 = -2.0 * coef * tsi * satrec.j3oj2 * satrec.meanMotion * sinio / satrec.ecco
		}
		satrec.x1mth2 = 1.0 - cosio2
		satrec.cc4 = 2.0 * satrec.meanMotion * coef1 * ao * omeosq * (satrec.eta*(2.0+0.5*etasq) + satrec.ecco*(0.5+2.0*etasq) - satrec.j2*tsi/(ao*psisq)*(-3.0*satrec.con41*(1.0-2.0*eeta+etasq*(1.5-0.5*eeta))+0.75*satrec.x1mth2*(2.0*etasq-eeta*(1.0+etasq))*math.Cos(2.0*satrec.argpoRad)))
		satrec.cc5 = 2.0 * coef1 * ao * omeosq * (1.0 + 2.75*(etasq+eeta) + eeta*etasq)
		cosio4 = cosio2 * cosio2
		temp1 = 1.5 * satrec.j2 * pinvsq * satrec.meanMotion
		temp2 = 0.5 * temp1 * satrec.j2 * pinvsq
		temp3 = -0.46875 * satrec.j4 * pinvsq * pinvsq * satrec.meanMotion
		satrec.mdot = satrec.meanMotion + 0.5*temp1*rteosq*satrec.con41 + 0.0625*temp2*rteosq*(13.0-78.0*cosio2+137.0*cosio4)
		satrec.argpdot = -0.5*temp1*con42 + 0.0625*temp2*(7.0-114.0*cosio2+395.0*cosio4) + temp3*(3.0-36.0*cosio2+49.0*cosio4)
		xhdot1 = -temp1 * cosio
		satrec.nodedot = xhdot1 + (0.5*temp2*(4.0-19.0*cosio2)+2.0*temp3*(3.0-7.0*cosio2))*cosio
		xpidot = satrec.argpdot + satrec.nodedot
		satrec.omgcof = satrec.bstar * cc3 * math.Cos(satrec.argpoRad)
		satrec.xmcof = 0.0
		if satrec.ecco > 1.0e-4 {
			satrec.xmcof = -x2o3 * coef * satrec.bstar / eeta
		}
		satrec.nodecf = 3.5 * omeosq * xhdot1 * satrec.cc1
		satrec.t2cof = 1.5 * satrec.cc1
		if math.Abs(cosio+1.0) > 1.5e-12 {
			satrec.xlcof = -0.25 * satrec.j3oj2 * sinio * (3.0 + 5.0*cosio) / (1.0 + cosio)
		} else {
			satrec.xlcof = -0.25 * satrec.j3oj2 * sinio * (3.0 + 5.0*cosio) / temp4
		}
		satrec.aycof = -0.5 * satrec.j3oj2 * sinio
		delmotemp = 1.0 + satrec.eta*math.Cos(satrec.moRad)
		satrec.delmo = delmotemp * delmotemp * delmotemp
		satrec.sinmao = math.Sin(satrec.moRad)
		satrec.x7thm1 = 7.0*cosio2 - 1.0
		if 2*3.14159265358979323846/satrec.meanMotion >= 225.0 {
			satrec.method = 'd'
			satrec.isimp = 1
			tc = 0.0
			inclmRad = satrec.incloRad
			dscom(epoch, satrec.ecco, satrec.argpoRad, tc, satrec.incloRad, satrec.nodeoRad, satrec.meanMotion, &snodm, &cnodm, &sinim, &cosim, &sinomm, &cosomm, &day, &satrec.e3, &satrec.ee2, &em, &emsq, &gamRad, &satrec.peo, &satrec.pgho, &satrec.pho, &satrec.pinco, &satrec.plo, &rtemsq, &satrec.se2, &satrec.se3, &satrec.sgh2, &satrec.sgh3, &satrec.sgh4, &satrec.sh2, &satrec.sh3, &satrec.si2, &satrec.si3, &satrec.sl2, &satrec.sl3, &satrec.sl4, &s1, &s2, &s3, &s4, &s5, &s6, &s7, &ss1, &ss2, &ss3, &ss4, &ss5, &ss6, &ss7, &sz1, &sz2, &sz3, &sz11, &sz12, &sz13, &sz21, &sz22, &sz23, &sz31, &sz32, &sz33, &satrec.xgh2, &satrec.xgh3, &satrec.xgh4, &satrec.xh2, &satrec.xh3, &satrec.xi2, &satrec.xi3, &satrec.xl2, &satrec.xl3, &satrec.xl4, &nm, &z1, &z2, &z3, &z11, &z12, &z13, &z21, &z22, &z23, &z31, &z32, &z33, &satrec.zmolRad, &satrec.zmosRad)
			dpper(satrec.e3, satrec.ee2, satrec.peo, satrec.pgho, satrec.pho, satrec.pinco, satrec.plo, satrec.se2, satrec.se3, satrec.sgh2, satrec.sgh3, satrec.sgh4, satrec.sh2, satrec.sh3, satrec.si2, satrec.si3, satrec.sl2, satrec.sl3, satrec.sl4, satrec.t, satrec.xgh2, satrec.xgh3, satrec.xgh4, satrec.xh2, satrec.xh3, satrec.xi2, satrec.xi3, satrec.xl2, satrec.xl3, satrec.xl4, satrec.zmolRad, satrec.zmosRad, inclmRad, satrec.init, &satrec.ecco, &satrec.incloRad, &satrec.nodeoRad, &satrec.argpoRad, &satrec.moRad, satrec.mode)
			argpmRad = 0.0
			nodemRad = 0.0
			mmRad = 0.0
			dsinit(satrec.xke, cosim, emsq, satrec.argpoRad, s1, s2, s3, s4, s5, sinim, ss1, ss2, ss3, ss4, ss5, sz1, sz3, sz11, sz13, sz21, sz23, sz31, sz33, satrec.t, tc, satrec.gstoRad, satrec.moRad, satrec.mdot, satrec.meanMotion, satrec.nodeoRad, satrec.nodedot, xpidot, z1, z3, z11, z13, z21, z23, z31, z33, satrec.ecco, eccsq, &em, &argpmRad, &inclmRad, &mmRad, &nm, &nodemRad, &satrec.irez, &satrec.atime, &satrec.d2201, &satrec.d2211, &satrec.d3210, &satrec.d3222, &satrec.d4410, &satrec.d4422, &satrec.d5220, &satrec.d5232, &satrec.d5421, &satrec.d5433, &satrec.dedt, &satrec.didt, &satrec.dmdt, &dndt, &satrec.dnodt, &satrec.domdt, &satrec.del1, &satrec.del2, &satrec.del3, &satrec.xfact, &satrec.xlamoRad, &satrec.xliRad, &satrec.xni)
		}
		if satrec.isimp != 1 {
			cc1sq = satrec.cc1 * satrec.cc1
			satrec.d2 = 4.0 * ao * tsi * cc1sq
			temp = satrec.d2 * tsi * satrec.cc1 / 3.0
			satrec.d3 = (17.0*ao + sfour) * temp
			satrec.d4 = 0.5 * temp * ao * tsi * (221.0*ao + 31.0*sfour) * satrec.cc1
			satrec.t3cof = satrec.d2 + 2.0*cc1sq
			satrec.t4cof = 0.25 * (3.0*satrec.d3 + satrec.cc1*(12.0*satrec.d2+10.0*cc1sq))
			satrec.t5cof = 0.2 * (3.0*satrec.d4 + 12.0*satrec.cc1*satrec.d3 + 6.0*satrec.d2*satrec.d2 + 15.0*cc1sq*(2.0*satrec.d2+cc1sq))
		}
	}
	propagate(satrec, &satrec.propagationState, 0.0, &r, &v)
	satrec.init = 'n'
	return true
}

func propagate(satrec *Record, work *propagationState, tsince float64, r *[3]float64, v *[3]float64) bool {
	var am, axnl, aynl, betal, cosim, cnod, cos2u, coseo1, cosi, cosip, cosisq, cossu, cosu, delm, delomg, em, emsq, ecose, el2, eo1Rad, ep, esine, argpmRad, argppRad, argpdfRad, pl, mrt, mvt, rdotl, rl, rvdot, rvdotl, sinim, sin2u, sineo1, sini, sinip, sinsu, sinu, snod, suRad, t2, t3, t4, tem5, temp, temp1, temp2, tempa, tempe, templ, uRad, ux, uy, uz, vx, vy, vz, inclmRad, mmRad, nm, nodemRad, xincRad, xincpRad, xlRad, xlmRad, mpRad, xmdfRad, xmx, xmy, nodedfRad, xnodeRad, nodepRad, tc, dndt, twopi, x2o3, vkmpersec, delmtemp, temp4 float64
	var ktr int
	mrt = 0.0
	temp4 = 1.5e-12
	twopi = 2.0 * 3.14159265358979323846
	x2o3 = 2.0 / 3.0
	vkmpersec = satrec.earthRadiusKM * satrec.xke / 60.0
	work.t = tsince
	work.error = 0
	xmdfRad = satrec.moRad + satrec.mdot*work.t
	argpdfRad = satrec.argpoRad + satrec.argpdot*work.t
	nodedfRad = satrec.nodeoRad + satrec.nodedot*work.t
	argpmRad = argpdfRad
	mmRad = xmdfRad
	t2 = work.t * work.t
	nodemRad = nodedfRad + satrec.nodecf*t2
	tempa = 1.0 - satrec.cc1*work.t
	tempe = satrec.bstar * satrec.cc4 * work.t
	templ = satrec.t2cof * t2
	if satrec.isimp != 1 {
		delomg = satrec.omgcof * work.t
		delmtemp = 1.0 + satrec.eta*math.Cos(xmdfRad)
		delm = satrec.xmcof * (delmtemp*delmtemp*delmtemp - satrec.delmo)
		temp = delomg + delm
		mmRad = xmdfRad + temp
		argpmRad = argpdfRad - temp
		t3 = t2 * work.t
		t4 = t3 * work.t
		tempa = tempa - satrec.d2*t2 - satrec.d3*t3 - satrec.d4*t4
		tempe = tempe + satrec.bstar*satrec.cc5*(math.Sin(mmRad)-satrec.sinmao)
		templ = templ + satrec.t3cof*t3 + t4*(satrec.t4cof+work.t*satrec.t5cof)
	}
	nm = satrec.meanMotion
	em = satrec.ecco
	inclmRad = satrec.incloRad
	if satrec.method == 'd' {
		tc = work.t
		dspace(satrec.irez, satrec.d2201, satrec.d2211, satrec.d3210, satrec.d3222, satrec.d4410, satrec.d4422, satrec.d5220, satrec.d5232, satrec.d5421, satrec.d5433, satrec.dedt, satrec.del1, satrec.del2, satrec.del3, satrec.didt, satrec.dmdt, satrec.dnodt, satrec.domdt, satrec.argpoRad, satrec.argpdot, work.t, tc, satrec.gstoRad, satrec.xfact, satrec.xlamoRad, satrec.meanMotion, &work.atime, &em, &argpmRad, &inclmRad, &work.xliRad, &mmRad, &work.xni, &nodemRad, &dndt, &nm)
	}
	if nm <= 0.0 {
		work.error = 2
		return false
	}
	am = math.Pow(satrec.xke/nm, x2o3) * tempa * tempa
	nm = satrec.xke / math.Pow(am, 1.5)
	em = em - tempe
	if em >= 1.0 || em < -0.001 {
		work.error = 1
		return false
	}
	if em < 1.0e-6 {
		em = 1.0e-6
	}
	mmRad = mmRad + satrec.meanMotion*templ
	xlmRad = mmRad + argpmRad + nodemRad
	emsq = em * em
	temp = 1.0 - emsq
	nodemRad = math.Mod(nodemRad, twopi)
	argpmRad = math.Mod(argpmRad, twopi)
	xlmRad = math.Mod(xlmRad, twopi)
	mmRad = math.Mod(xlmRad-argpmRad-nodemRad, twopi)
	work.am = am
	work.em = em
	work.inclinationMeanRad = inclmRad
	work.nodeMeanRad = nodemRad
	work.om = argpmRad
	work.mmRad = mmRad
	work.nm = nm
	sinim = math.Sin(inclmRad)
	cosim = math.Cos(inclmRad)
	ep = em
	xincpRad = inclmRad
	argppRad = argpmRad
	nodepRad = nodemRad
	mpRad = mmRad
	sinip = sinim
	cosip = cosim
	if satrec.method == 'd' {
		dpper(satrec.e3, satrec.ee2, satrec.peo, satrec.pgho, satrec.pho, satrec.pinco, satrec.plo, satrec.se2, satrec.se3, satrec.sgh2, satrec.sgh3, satrec.sgh4, satrec.sh2, satrec.sh3, satrec.si2, satrec.si3, satrec.sl2, satrec.sl3, satrec.sl4, work.t, satrec.xgh2, satrec.xgh3, satrec.xgh4, satrec.xh2, satrec.xh3, satrec.xi2, satrec.xi3, satrec.xl2, satrec.xl3, satrec.xl4, satrec.zmolRad, satrec.zmosRad, satrec.incloRad, 'n', &ep, &xincpRad, &nodepRad, &argppRad, &mpRad, satrec.mode)
		if xincpRad < 0.0 {
			xincpRad = -xincpRad
			nodepRad = nodepRad + 3.14159265358979323846
			argppRad = argppRad - 3.14159265358979323846
		}
		if ep < 0.0 || ep > 1.0 {
			work.error = 3
			return false
		}
	}
	if satrec.method == 'd' {
		sinip = math.Sin(xincpRad)
		cosip = math.Cos(xincpRad)
		work.aycof = -0.5 * satrec.j3oj2 * sinip
		if math.Abs(cosip+1.0) > 1.5e-12 {
			work.xlcof = -0.25 * satrec.j3oj2 * sinip * (3.0 + 5.0*cosip) / (1.0 + cosip)
		} else {
			work.xlcof = -0.25 * satrec.j3oj2 * sinip * (3.0 + 5.0*cosip) / temp4
		}
	}
	axnl = ep * math.Cos(argppRad)
	temp = 1.0 / (am * (1.0 - ep*ep))
	aynl = ep*math.Sin(argppRad) + temp*work.aycof
	xlRad = mpRad + argppRad + nodepRad + temp*work.xlcof*axnl
	uRad = math.Mod(xlRad-nodepRad, twopi)
	eo1Rad = uRad
	tem5 = 9999.9
	ktr = 1
	for math.Abs(tem5) >= 1.0e-12 && ktr <= 10 {
		sineo1 = math.Sin(eo1Rad)
		coseo1 = math.Cos(eo1Rad)
		tem5 = 1.0 - coseo1*axnl - sineo1*aynl
		tem5 = (uRad - aynl*coseo1 + axnl*sineo1 - eo1Rad) / tem5
		if math.Abs(tem5) >= 0.95 {
			tem5 = func() float64 {
				if tem5 > 0.0 {
					return 0.95
				}
				return -0.95
			}()
		}
		eo1Rad = eo1Rad + tem5
		ktr = ktr + 1
	}
	ecose = axnl*coseo1 + aynl*sineo1
	esine = axnl*sineo1 - aynl*coseo1
	el2 = axnl*axnl + aynl*aynl
	pl = am * (1.0 - el2)
	if pl < 0.0 {
		work.error = 4
		return false
	} else {
		rl = am * (1.0 - ecose)
		rdotl = math.Sqrt(am) * esine / rl
		rvdotl = math.Sqrt(pl) / rl
		betal = math.Sqrt(1.0 - el2)
		temp = esine / (1.0 + betal)
		sinu = am / rl * (sineo1 - aynl - axnl*temp)
		cosu = am / rl * (coseo1 - axnl + aynl*temp)
		suRad = math.Atan2(sinu, cosu)
		sin2u = (cosu + cosu) * sinu
		cos2u = 1.0 - 2.0*sinu*sinu
		temp = 1.0 / pl
		temp1 = 0.5 * satrec.j2 * temp
		temp2 = temp1 * temp
		if satrec.method == 'd' {
			cosisq = cosip * cosip
			work.con41 = 3.0*cosisq - 1.0
			work.x1mth2 = 1.0 - cosisq
			work.x7thm1 = 7.0*cosisq - 1.0
		}
		mrt = rl*(1.0-1.5*temp2*betal*work.con41) + 0.5*temp1*work.x1mth2*cos2u
		suRad = suRad - 0.25*temp2*work.x7thm1*sin2u
		xnodeRad = nodepRad + 1.5*temp2*cosip*sin2u
		xincRad = xincpRad + 1.5*temp2*cosip*sinip*cos2u
		mvt = rdotl - nm*temp1*work.x1mth2*sin2u/satrec.xke
		rvdot = rvdotl + nm*temp1*(work.x1mth2*cos2u+1.5*work.con41)/satrec.xke
		sinsu = math.Sin(suRad)
		cossu = math.Cos(suRad)
		snod = math.Sin(xnodeRad)
		cnod = math.Cos(xnodeRad)
		sini = math.Sin(xincRad)
		cosi = math.Cos(xincRad)
		xmx = -snod * cosi
		xmy = cnod * cosi
		ux = xmx*sinsu + cnod*cossu
		uy = xmy*sinsu + snod*cossu
		uz = sini * sinsu
		vx = xmx*cossu - cnod*sinsu
		vy = xmy*cossu - snod*sinsu
		vz = sini * cossu
		r[0] = mrt * ux * satrec.earthRadiusKM
		r[1] = mrt * uy * satrec.earthRadiusKM
		r[2] = mrt * uz * satrec.earthRadiusKM
		v[0] = (mvt*ux + rvdot*vx) * vkmpersec
		v[1] = (mvt*uy + rvdot*vy) * vkmpersec
		v[2] = (mvt*uz + rvdot*vz) * vkmpersec
	}
	if mrt < 1.0 {
		work.error = 6
		return false
	}
	return true
}
