package sgp4

import "math"

func dpper(e3 float64, ee2 float64, peo float64, pgho float64, pho float64, pinco float64, plo float64, se2 float64, se3 float64, sgh2 float64, sgh3 float64, sgh4 float64, sh2 float64, sh3 float64, si2 float64, si3 float64, sl2 float64, sl3 float64, sl4 float64, t float64, xgh2 float64, xgh3 float64, xgh4 float64, xh2 float64, xh3 float64, xi2 float64, xi3 float64, xl2 float64, xl3 float64, xl4 float64, zmolRad float64, zmosRad float64, incloRad float64, init byte, ep *float64, inclpRad *float64, nodepRad *float64, argppRad *float64, mpRad *float64, opsmode byte) {
	var twopi, alfdp, betdp, cosip, cosop, dalf, dbet, dls, f2, f3, pe, pgh, ph, pinc, pl, sel, ses, sghl, sghs, shll, shs, sil, sinip, sinop, sinzf, sis, sll, sls, xlsRad, xnohRad, zfRad, zmRad, zel, zes, znl, zns float64
	twopi = 2.0 * 3.14159265358979323846
	zns = 1.19459e-5
	zes = 0.01675
	znl = 1.5835218e-4
	zel = 0.05490
	zmRad = zmosRad + zns*t
	if init == 'y' {
		zmRad = zmosRad
	}
	zfRad = zmRad + 2.0*zes*math.Sin(zmRad)
	sinzf = math.Sin(zfRad)
	f2 = 0.5*sinzf*sinzf - 0.25
	f3 = -0.5 * sinzf * math.Cos(zfRad)
	ses = se2*f2 + se3*f3
	sis = si2*f2 + si3*f3
	sls = sl2*f2 + sl3*f3 + sl4*sinzf
	sghs = sgh2*f2 + sgh3*f3 + sgh4*sinzf
	shs = sh2*f2 + sh3*f3
	zmRad = zmolRad + znl*t
	if init == 'y' {
		zmRad = zmolRad
	}
	zfRad = zmRad + 2.0*zel*math.Sin(zmRad)
	sinzf = math.Sin(zfRad)
	f2 = 0.5*sinzf*sinzf - 0.25
	f3 = -0.5 * sinzf * math.Cos(zfRad)
	sel = ee2*f2 + e3*f3
	sil = xi2*f2 + xi3*f3
	sll = xl2*f2 + xl3*f3 + xl4*sinzf
	sghl = xgh2*f2 + xgh3*f3 + xgh4*sinzf
	shll = xh2*f2 + xh3*f3
	pe = ses + sel
	pinc = sis + sil
	pl = sls + sll
	pgh = sghs + sghl
	ph = shs + shll
	if init == 'n' {
		pe = pe - peo
		pinc = pinc - pinco
		pl = pl - plo
		pgh = pgh - pgho
		ph = ph - pho
		*inclpRad = *inclpRad + pinc
		*ep = *ep + pe
		sinip = math.Sin(*inclpRad)
		cosip = math.Cos(*inclpRad)
		if *inclpRad >= 0.2 {
			ph = ph / sinip
			pgh = pgh - cosip*ph
			*argppRad = *argppRad + pgh
			*nodepRad = *nodepRad + ph
			*mpRad = *mpRad + pl
		} else {
			sinop = math.Sin(*nodepRad)
			cosop = math.Cos(*nodepRad)
			alfdp = sinip * sinop
			betdp = sinip * cosop
			dalf = ph*cosop + pinc*cosip*sinop
			dbet = -ph*sinop + pinc*cosip*cosop
			alfdp = alfdp + dalf
			betdp = betdp + dbet
			*nodepRad = math.Mod(*nodepRad, twopi)
			if *nodepRad < 0.0 && opsmode == 'a' {
				*nodepRad = *nodepRad + twopi
			}
			xlsRad = *mpRad + *argppRad + cosip**nodepRad
			dls = pl + pgh - pinc**nodepRad*sinip
			xlsRad = xlsRad + dls
			xnohRad = *nodepRad
			*nodepRad = math.Atan2(alfdp, betdp)
			if *nodepRad < 0.0 && opsmode == 'a' {
				*nodepRad = *nodepRad + twopi
			}
			if math.Abs(xnohRad-*nodepRad) > 3.14159265358979323846 {
				if *nodepRad < xnohRad {
					*nodepRad = *nodepRad + twopi
				} else {
					*nodepRad = *nodepRad - twopi
				}
			}
			*mpRad = *mpRad + pl
			*argppRad = xlsRad - *mpRad - cosip**nodepRad
		}
	}
}

func dscom(epoch float64, ep float64, argppRad float64, tc float64, inclpRad float64, nodepRad float64, np float64, snodm *float64, cnodm *float64, sinim *float64, cosim *float64, sinomm *float64, cosomm *float64, day *float64, e3 *float64, ee2 *float64, em *float64, emsq *float64, gamRad *float64, peo *float64, pgho *float64, pho *float64, pinco *float64, plo *float64, rtemsq *float64, se2 *float64, se3 *float64, sgh2 *float64, sgh3 *float64, sgh4 *float64, sh2 *float64, sh3 *float64, si2 *float64, si3 *float64, sl2 *float64, sl3 *float64, sl4 *float64, s1 *float64, s2 *float64, s3 *float64, s4 *float64, s5 *float64, s6 *float64, s7 *float64, ss1 *float64, ss2 *float64, ss3 *float64, ss4 *float64, ss5 *float64, ss6 *float64, ss7 *float64, sz1 *float64, sz2 *float64, sz3 *float64, sz11 *float64, sz12 *float64, sz13 *float64, sz21 *float64, sz22 *float64, sz23 *float64, sz31 *float64, sz32 *float64, sz33 *float64, xgh2 *float64, xgh3 *float64, xgh4 *float64, xh2 *float64, xh3 *float64, xi2 *float64, xi3 *float64, xl2 *float64, xl3 *float64, xl4 *float64, nm *float64, z1 *float64, z2 *float64, z3 *float64, z11 *float64, z12 *float64, z13 *float64, z21 *float64, z22 *float64, z23 *float64, z31 *float64, z32 *float64, z33 *float64, zmolRad *float64, zmosRad *float64) {
	var zes, zel, c1ss, c1l, zsinis, zcosis, zcosgs, zsings, twopi, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, betasq, cc, ctem, stem, x1, x2, x3, x4, x5, x6, x7, x8, xnodceRad, xnoi, zcosg, zcosgl, zcosh, zcoshl, zcosi, zcosil, zsing, zsingl, zsinh, zsinhl, zsini, zsinil, zxRad, zyRad float64
	var lsflg int
	zes = 0.01675
	zel = 0.05490
	c1ss = 2.9864797e-6
	c1l = 4.7968065e-7
	zsinis = 0.39785416
	zcosis = 0.91744867
	zcosgs = 0.1945905
	zsings = -0.98088458
	twopi = 2.0 * 3.14159265358979323846
	*nm = np
	*em = ep
	*snodm = math.Sin(nodepRad)
	*cnodm = math.Cos(nodepRad)
	*sinomm = math.Sin(argppRad)
	*cosomm = math.Cos(argppRad)
	*sinim = math.Sin(inclpRad)
	*cosim = math.Cos(inclpRad)
	*emsq = *em * *em
	betasq = 1.0 - *emsq
	*rtemsq = math.Sqrt(betasq)
	*peo = 0.0
	*pinco = 0.0
	*plo = 0.0
	*pgho = 0.0
	*pho = 0.0
	*day = epoch + 18261.5 + tc/1440.0
	xnodceRad = math.Mod(4.5236020-9.2422029e-4**day, twopi)
	stem = math.Sin(xnodceRad)
	ctem = math.Cos(xnodceRad)
	zcosil = 0.91375164 - 0.03568096*ctem
	zsinil = math.Sqrt(1.0 - zcosil*zcosil)
	zsinhl = 0.089683511 * stem / zsinil
	zcoshl = math.Sqrt(1.0 - zsinhl*zsinhl)
	*gamRad = 5.8351514 + 0.0019443680**day
	zxRad = 0.39785416 * stem / zsinil
	zyRad = zcoshl*ctem + 0.91744867*zsinhl*stem
	zxRad = math.Atan2(zxRad, zyRad)
	zxRad = *gamRad + zxRad - xnodceRad
	zcosgl = math.Cos(zxRad)
	zsingl = math.Sin(zxRad)
	zcosg = zcosgs
	zsing = zsings
	zcosi = zcosis
	zsini = zsinis
	zcosh = *cnodm
	zsinh = *snodm
	cc = c1ss
	xnoi = 1.0 / *nm
	lsflg = 1
	for lsflg <= 2 {
		a1 = zcosg*zcosh + zsing*zcosi*zsinh
		a3 = -zsing*zcosh + zcosg*zcosi*zsinh
		a7 = -zcosg*zsinh + zsing*zcosi*zcosh
		a8 = zsing * zsini
		a9 = zsing*zsinh + zcosg*zcosi*zcosh
		a10 = zcosg * zsini
		a2 = *cosim*a7 + *sinim*a8
		a4 = *cosim*a9 + *sinim*a10
		a5 = -*sinim*a7 + *cosim*a8
		a6 = -*sinim*a9 + *cosim*a10
		x1 = a1**cosomm + a2**sinomm
		x2 = a3**cosomm + a4**sinomm
		x3 = -a1**sinomm + a2**cosomm
		x4 = -a3**sinomm + a4**cosomm
		x5 = a5 * *sinomm
		x6 = a6 * *sinomm
		x7 = a5 * *cosomm
		x8 = a6 * *cosomm
		*z31 = 12.0*x1*x1 - 3.0*x3*x3
		*z32 = 24.0*x1*x2 - 6.0*x3*x4
		*z33 = 12.0*x2*x2 - 3.0*x4*x4
		*z1 = 3.0*(a1*a1+a2*a2) + *z31**emsq
		*z2 = 6.0*(a1*a3+a2*a4) + *z32**emsq
		*z3 = 3.0*(a3*a3+a4*a4) + *z33**emsq
		*z11 = -6.0*a1*a5 + *emsq*(-24.0*x1*x7-6.0*x3*x5)
		*z12 = -6.0*(a1*a6+a3*a5) + *emsq*(-24.0*(x2*x7+x1*x8)-6.0*(x3*x6+x4*x5))
		*z13 = -6.0*a3*a6 + *emsq*(-24.0*x2*x8-6.0*x4*x6)
		*z21 = 6.0*a2*a5 + *emsq*(24.0*x1*x5-6.0*x3*x7)
		*z22 = 6.0*(a4*a5+a2*a6) + *emsq*(24.0*(x2*x5+x1*x6)-6.0*(x4*x7+x3*x8))
		*z23 = 6.0*a4*a6 + *emsq*(24.0*x2*x6-6.0*x4*x8)
		*z1 = *z1 + *z1 + betasq**z31
		*z2 = *z2 + *z2 + betasq**z32
		*z3 = *z3 + *z3 + betasq**z33
		*s3 = cc * xnoi
		*s2 = -0.5 * *s3 / *rtemsq
		*s4 = *s3 * *rtemsq
		*s1 = -15.0 * *em * *s4
		*s5 = x1*x3 + x2*x4
		*s6 = x2*x3 + x1*x4
		*s7 = x2*x4 - x1*x3
		if lsflg == 1 {
			*ss1 = *s1
			*ss2 = *s2
			*ss3 = *s3
			*ss4 = *s4
			*ss5 = *s5
			*ss6 = *s6
			*ss7 = *s7
			*sz1 = *z1
			*sz2 = *z2
			*sz3 = *z3
			*sz11 = *z11
			*sz12 = *z12
			*sz13 = *z13
			*sz21 = *z21
			*sz22 = *z22
			*sz23 = *z23
			*sz31 = *z31
			*sz32 = *z32
			*sz33 = *z33
			zcosg = zcosgl
			zsing = zsingl
			zcosi = zcosil
			zsini = zsinil
			zcosh = zcoshl**cnodm + zsinhl**snodm
			zsinh = *snodm*zcoshl - *cnodm*zsinhl
			cc = c1l
		}
		lsflg++
	}
	*zmolRad = math.Mod(4.7199672+0.22997150**day-*gamRad, twopi)
	*zmosRad = math.Mod(6.2565837+0.017201977**day, twopi)
	*se2 = 2.0 * *ss1 * *ss6
	*se3 = 2.0 * *ss1 * *ss7
	*si2 = 2.0 * *ss2 * *sz12
	*si3 = 2.0 * *ss2 * (*sz13 - *sz11)
	*sl2 = -2.0 * *ss3 * *sz2
	*sl3 = -2.0 * *ss3 * (*sz3 - *sz1)
	*sl4 = -2.0 * *ss3 * (-21.0 - 9.0**emsq) * zes
	*sgh2 = 2.0 * *ss4 * *sz32
	*sgh3 = 2.0 * *ss4 * (*sz33 - *sz31)
	*sgh4 = -18.0 * *ss4 * zes
	*sh2 = -2.0 * *ss2 * *sz22
	*sh3 = -2.0 * *ss2 * (*sz23 - *sz21)
	*ee2 = 2.0 * *s1 * *s6
	*e3 = 2.0 * *s1 * *s7
	*xi2 = 2.0 * *s2 * *z12
	*xi3 = 2.0 * *s2 * (*z13 - *z11)
	*xl2 = -2.0 * *s3 * *z2
	*xl3 = -2.0 * *s3 * (*z3 - *z1)
	*xl4 = -2.0 * *s3 * (-21.0 - 9.0**emsq) * zel
	*xgh2 = 2.0 * *s4 * *z32
	*xgh3 = 2.0 * *s4 * (*z33 - *z31)
	*xgh4 = -18.0 * *s4 * zel
	*xh2 = -2.0 * *s2 * *z22
	*xh3 = -2.0 * *s2 * (*z23 - *z21)
}

func dsinit(xke float64, cosim float64, emsq float64, argpoRad float64, s1 float64, s2 float64, s3 float64, s4 float64, s5 float64, sinim float64, ss1 float64, ss2 float64, ss3 float64, ss4 float64, ss5 float64, sz1 float64, sz3 float64, sz11 float64, sz13 float64, sz21 float64, sz23 float64, sz31 float64, sz33 float64, t float64, tc float64, gstoRad float64, moRad float64, mdot float64, no float64, nodeoRad float64, nodedot float64, xpidot float64, z1 float64, z3 float64, z11 float64, z13 float64, z21 float64, z23 float64, z31 float64, z33 float64, ecco float64, eccsq float64, em *float64, argpmRad *float64, inclmRad *float64, mmRad *float64, nm *float64, nodemRad *float64, irez *int, atime *float64, d2201 *float64, d2211 *float64, d3210 *float64, d3222 *float64, d4410 *float64, d4422 *float64, d5220 *float64, d5232 *float64, d5421 *float64, d5433 *float64, dedt *float64, didt *float64, dmdt *float64, dndt *float64, dnodt *float64, domdt *float64, del1 *float64, del2 *float64, del3 *float64, xfact *float64, xlamoRad *float64, xliRad *float64, xni *float64) {
	var twopi, ainv2, aonv, cosisq, eoc, f220, f221, f311, f321, f322, f330, f441, f442, f522, f523, f542, f543, g200, g201, g211, g300, g310, g322, g410, g422, g520, g521, g532, g533, ses, sgs, sghl, sghs, shs, shll, sis, sini2, sls, temp, temp1, thetaRad, xno2, q22, q31, q33, root22, root44, root54, rptim, root32, root52, x2o3, znl, emo, zns, emsqo float64
	twopi = 2.0 * 3.14159265358979323846
	aonv = 0.0
	q22 = 1.7891679e-6
	q31 = 2.1460748e-6
	q33 = 2.2123015e-7
	root22 = 1.7891679e-6
	root44 = 7.3636953e-9
	root54 = 2.1765803e-9
	rptim = 4.37526908801129966e-3
	root32 = 3.7393792e-7
	root52 = 1.1428639e-7
	x2o3 = 2.0 / 3.0
	znl = 1.5835218e-4
	zns = 1.19459e-5
	*irez = 0
	if *nm < 0.0052359877 && *nm > 0.0034906585 {
		*irez = 1
	}
	if *nm >= 8.26e-3 && *nm <= 9.24e-3 && *em >= 0.5 {
		*irez = 2
	}
	ses = ss1 * zns * ss5
	sis = ss2 * zns * (sz11 + sz13)
	sls = -zns * ss3 * (sz1 + sz3 - 14.0 - 6.0*emsq)
	sghs = ss4 * zns * (sz31 + sz33 - 6.0)
	shs = -zns * ss2 * (sz21 + sz23)
	if *inclmRad < 5.2359877e-2 || *inclmRad > 3.14159265358979323846-5.2359877e-2 {
		shs = 0.0
	}
	if math.Abs(sinim) > 0.0 {
		shs = shs / sinim
	}
	sgs = sghs - cosim*shs
	*dedt = ses + s1*znl*s5
	*didt = sis + s2*znl*(z11+z13)
	*dmdt = sls - znl*s3*(z1+z3-14.0-6.0*emsq)
	sghl = s4 * znl * (z31 + z33 - 6.0)
	shll = -znl * s2 * (z21 + z23)
	if *inclmRad < 5.2359877e-2 || *inclmRad > 3.14159265358979323846-5.2359877e-2 {
		shll = 0.0
	}
	*domdt = sgs + sghl
	*dnodt = shs
	if math.Abs(sinim) > 0.0 {
		*domdt = *domdt - cosim/sinim*shll
		*dnodt = *dnodt + shll/sinim
	}
	*dndt = 0.0
	thetaRad = math.Mod(gstoRad+tc*rptim, twopi)
	*em = *em + *dedt*t
	*inclmRad = *inclmRad + *didt*t
	*argpmRad = *argpmRad + *domdt*t
	*nodemRad = *nodemRad + *dnodt*t
	*mmRad = *mmRad + *dmdt*t
	if *irez != 0 {
		aonv = math.Pow(*nm/xke, x2o3)
		if *irez == 2 {
			cosisq = cosim * cosim
			emo = *em
			*em = ecco
			emsqo = emsq
			emsq = eccsq
			eoc = *em * emsq
			g201 = -0.306 - (*em-0.64)*0.440
			if *em <= 0.65 {
				g211 = 3.616 - 13.2470**em + 16.2900*emsq
				g310 = -19.302 + 117.3900**em - 228.4190*emsq + 156.5910*eoc
				g322 = -18.9068 + 109.7927**em - 214.6334*emsq + 146.5816*eoc
				g410 = -41.122 + 242.6940**em - 471.0940*emsq + 313.9530*eoc
				g422 = -146.407 + 841.8800**em - 1629.014*emsq + 1083.4350*eoc
				g520 = -532.114 + 3017.977**em - 5740.032*emsq + 3708.2760*eoc
			} else {
				g211 = -72.099 + 331.819**em - 508.738*emsq + 266.724*eoc
				g310 = -346.844 + 1582.851**em - 2415.925*emsq + 1246.113*eoc
				g322 = -342.585 + 1554.908**em - 2366.899*emsq + 1215.972*eoc
				g410 = -1052.797 + 4758.686**em - 7193.992*emsq + 3651.957*eoc
				g422 = -3581.690 + 16178.110**em - 24462.770*emsq + 12422.520*eoc
				if *em > 0.715 {
					g520 = -5149.66 + 29936.92**em - 54087.36*emsq + 31324.56*eoc
				} else {
					g520 = 1464.74 - 4664.75**em + 3763.64*emsq
				}
			}
			if *em < 0.7 {
				g533 = -919.22770 + 4988.6100**em - 9064.7700*emsq + 5542.21*eoc
				g521 = -822.71072 + 4568.6173**em - 8491.4146*emsq + 5337.524*eoc
				g532 = -853.66600 + 4690.2500**em - 8624.7700*emsq + 5341.4*eoc
			} else {
				g533 = -37995.780 + 161616.52**em - 229838.20*emsq + 109377.94*eoc
				g521 = -51752.104 + 218913.95**em - 309468.16*emsq + 146349.42*eoc
				g532 = -40023.880 + 170470.89**em - 242699.48*emsq + 115605.82*eoc
			}
			sini2 = sinim * sinim
			f220 = 0.75 * (1.0 + 2.0*cosim + cosisq)
			f221 = 1.5 * sini2
			f321 = 1.875 * sinim * (1.0 - 2.0*cosim - 3.0*cosisq)
			f322 = -1.875 * sinim * (1.0 + 2.0*cosim - 3.0*cosisq)
			f441 = 35.0 * sini2 * f220
			f442 = 39.3750 * sini2 * sini2
			f522 = 9.84375 * sinim * (sini2*(1.0-2.0*cosim-5.0*cosisq) + 0.33333333*(-2.0+4.0*cosim+6.0*cosisq))
			f523 = sinim * (4.92187512*sini2*(-2.0-4.0*cosim+10.0*cosisq) + 6.56250012*(1.0+2.0*cosim-3.0*cosisq))
			f542 = 29.53125 * sinim * (2.0 - 8.0*cosim + cosisq*(-12.0+8.0*cosim+10.0*cosisq))
			f543 = 29.53125 * sinim * (-2.0 - 8.0*cosim + cosisq*(12.0+8.0*cosim-10.0*cosisq))
			xno2 = *nm * *nm
			ainv2 = aonv * aonv
			temp1 = 3.0 * xno2 * ainv2
			temp = temp1 * root22
			*d2201 = temp * f220 * g201
			*d2211 = temp * f221 * g211
			temp1 = temp1 * aonv
			temp = temp1 * root32
			*d3210 = temp * f321 * g310
			*d3222 = temp * f322 * g322
			temp1 = temp1 * aonv
			temp = 2.0 * temp1 * root44
			*d4410 = temp * f441 * g410
			*d4422 = temp * f442 * g422
			temp1 = temp1 * aonv
			temp = temp1 * root52
			*d5220 = temp * f522 * g520
			*d5232 = temp * f523 * g532
			temp = 2.0 * temp1 * root54
			*d5421 = temp * f542 * g521
			*d5433 = temp * f543 * g533
			*xlamoRad = math.Mod(moRad+nodeoRad+nodeoRad-thetaRad-thetaRad, twopi)
			*xfact = mdot + *dmdt + 2.0*(nodedot+*dnodt-rptim) - no
			*em = emo
			emsq = emsqo
		}
		if *irez == 1 {
			g200 = 1.0 + emsq*(-2.5+0.8125*emsq)
			g310 = 1.0 + 2.0*emsq
			g300 = 1.0 + emsq*(-6.0+6.60937*emsq)
			f220 = 0.75 * (1.0 + cosim) * (1.0 + cosim)
			f311 = 0.9375*sinim*sinim*(1.0+3.0*cosim) - 0.75*(1.0+cosim)
			f330 = 1.0 + cosim
			f330 = 1.875 * f330 * f330 * f330
			*del1 = 3.0 * *nm * *nm * aonv * aonv
			*del2 = 2.0 * *del1 * f220 * g200 * q22
			*del3 = 3.0 * *del1 * f330 * g300 * q33 * aonv
			*del1 = *del1 * f311 * g310 * q31 * aonv
			*xlamoRad = math.Mod(moRad+nodeoRad+argpoRad-thetaRad, twopi)
			*xfact = mdot + xpidot - rptim + *dmdt + *domdt + *dnodt - no
		}
		*xliRad = *xlamoRad
		*xni = no
		*atime = 0.0
		*nm = no + *dndt
	}
}

func dspace(irez int, d2201 float64, d2211 float64, d3210 float64, d3222 float64, d4410 float64, d4422 float64, d5220 float64, d5232 float64, d5421 float64, d5433 float64, dedt float64, del1 float64, del2 float64, del3 float64, didt float64, dmdt float64, dnodt float64, domdt float64, argpoRad float64, argpdot float64, t float64, tc float64, gstoRad float64, xfact float64, xlamoRad float64, no float64, atime *float64, em *float64, argpmRad *float64, inclmRad *float64, xliRad *float64, mmRad *float64, xni *float64, nodemRad *float64, dndt *float64, nm *float64) {
	var twopi, delt, ft, thetaRad, x2li, x2omi, xlRad, xldot, xnddt, xndt, xomi, g22, g32, g44, g52, g54, fasx2, fasx4, fasx6, rptim, step2, stepn, stepp float64
	var iretn int
	twopi = 2.0 * 3.14159265358979323846
	fasx2 = 0.13130908
	fasx4 = 2.8843198
	fasx6 = 0.37448087
	g22 = 5.7686396
	g32 = 0.95240898
	g44 = 1.8014998
	g52 = 1.0508330
	g54 = 4.4108898
	rptim = 4.37526908801129966e-3
	stepp = 720.0
	stepn = -720.0
	step2 = 259200.0
	*dndt = 0.0
	thetaRad = math.Mod(gstoRad+tc*rptim, twopi)
	*em = *em + dedt*t
	*inclmRad = *inclmRad + didt*t
	*argpmRad = *argpmRad + domdt*t
	*nodemRad = *nodemRad + dnodt*t
	*mmRad = *mmRad + dmdt*t
	ft = 0.0
	if irez != 0 {
		if math.Abs(*atime) <= 0.0 || t**atime <= 0.0 || math.Abs(t) < math.Abs(*atime) {
			*atime = 0.0
			*xni = no
			*xliRad = xlamoRad
		}
		if t > 0.0 {
			delt = stepp
		} else {
			delt = stepn
		}
		iretn = 381
		for iretn == 381 {
			if irez != 2 {
				xndt = del1*math.Sin(*xliRad-fasx2) + del2*math.Sin(2.0*(*xliRad-fasx4)) + del3*math.Sin(3.0*(*xliRad-fasx6))
				xldot = *xni + xfact
				xnddt = del1*math.Cos(*xliRad-fasx2) + 2.0*del2*math.Cos(2.0*(*xliRad-fasx4)) + 3.0*del3*math.Cos(3.0*(*xliRad-fasx6))
				xnddt = xnddt * xldot
			} else {
				xomi = argpoRad + argpdot**atime
				x2omi = xomi + xomi
				x2li = *xliRad + *xliRad
				xndt = d2201*math.Sin(x2omi+*xliRad-g22) + d2211*math.Sin(*xliRad-g22) + d3210*math.Sin(xomi+*xliRad-g32) + d3222*math.Sin(-xomi+*xliRad-g32) + d4410*math.Sin(x2omi+x2li-g44) + d4422*math.Sin(x2li-g44) + d5220*math.Sin(xomi+*xliRad-g52) + d5232*math.Sin(-xomi+*xliRad-g52) + d5421*math.Sin(xomi+x2li-g54) + d5433*math.Sin(-xomi+x2li-g54)
				xldot = *xni + xfact
				xnddt = d2201*math.Cos(x2omi+*xliRad-g22) + d2211*math.Cos(*xliRad-g22) + d3210*math.Cos(xomi+*xliRad-g32) + d3222*math.Cos(-xomi+*xliRad-g32) + d5220*math.Cos(xomi+*xliRad-g52) + d5232*math.Cos(-xomi+*xliRad-g52) + 2.0*(d4410*math.Cos(x2omi+x2li-g44)+d4422*math.Cos(x2li-g44)+d5421*math.Cos(xomi+x2li-g54)+d5433*math.Cos(-xomi+x2li-g54))
				xnddt = xnddt * xldot
			}
			if math.Abs(t-*atime) >= stepp {
				iretn = 381
			} else {
				ft = t - *atime
				iretn = 0
			}
			if iretn == 381 {
				*xliRad = *xliRad + xldot*delt + xndt*step2
				*xni = *xni + xndt*delt + xnddt*step2
				*atime = *atime + delt
			}
		}
		*nm = *xni + xndt*ft + xnddt*ft*ft*0.5
		xlRad = *xliRad + xldot*ft + xndt*ft*ft*0.5
		if irez != 1 {
			*mmRad = xlRad - 2.0**nodemRad + 2.0*thetaRad
			*dndt = *nm - no
		} else {
			*mmRad = xlRad - *nodemRad - *argpmRad + thetaRad
			*dndt = *nm - no
		}
		*nm = no + *dndt
	}
}
