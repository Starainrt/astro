package basic

import "math"

const (
	// 线性时标关系的参考历元，不表示各时标读数在此相等。
	barycentricT0JDE = 2443144.5003725
	lgRate           = 6.969290134e-10
	lbRate           = 1.550519768e-8
	tdb0Seconds      = -65.5e-6
)

// tdbPeriodicGroups 按 t 的幂次分组的 TDB−TT 周期项（振幅秒、频率 rad/儒略千年、相位 rad）。
var tdbPeriodicGroups = [5][][3]float64{
	{
		{0.001656674564, 6283.075849991, 6.240054195},
		{2.2417471e-05, 5753.384884897, 4.296977442},
		{1.3839792e-05, 12566.151699983, 6.19690441},
		{4.770086e-06, 529.690965095, 0.444401603},
		{4.67674e-06, 6069.776754553, 4.021195093},
		{2.256707e-06, 213.299095438, 5.543113262},
		{1.694205e-06, -3.523118349, 5.025132748},
		{1.554905e-06, 77713.77146792, 5.19846709},
		{1.276839e-06, 7860.419392439, 5.988822341},
		{1.193379e-06, 5223.693919802, 3.64982373},
		{1.115322e-06, 3930.20969622, 1.422745069},
		{7.94185e-07, 11506.769769794, 2.322313077},
		{4.47061e-07, 26.2983198, 3.615796498},
		{4.35206e-07, -398.149003408, 4.349338347},
		{6.00309e-07, 1577.343542448, 2.678271909},
		{4.96817e-07, 6208.294251424, 5.696701824},
		{4.86306e-07, 5884.926846583, 0.520007179},
		{4.32392e-07, 74.781598567, 2.435898309},
		{4.68597e-07, 6244.942814354, 5.866398759},
		{3.7551e-07, 5507.553238667, 4.103476804},
		{2.43085e-07, -775.522611324, 3.651837925},
		{1.73435e-07, 18849.227549974, 6.153743485},
		{2.30685e-07, 5856.477659115, 4.773852582},
		{2.03747e-07, 12036.460734888, 4.333987818},
	},
	{
		{0.000102156724, 6283.075849991, 4.249032005},
		{1.706807e-06, 12566.151699983, 4.205904248},
		{2.69668e-07, 213.299095438, 3.400290479},
		{2.65919e-07, 529.690965095, 5.836047367},
		{2.10568e-07, -3.523118349, 6.262738348},
		{7.7996e-08, 5223.693919802, 4.670344204},
		{5.4764e-08, 1577.343542448, 4.53480017},
		{5.9146e-08, 26.2983198, 1.083044735},
	},
	{
		{4.32299e-06, 6283.075849991, 2.642893748},
		{4.06495e-07, 0.0, 4.71238898},
		{1.22605e-07, 12566.151699983, 2.438140634},
		{1.9476e-08, 213.299095438, 1.642186981},
	},
	{
		{1.43388e-07, 6283.075849991, 1.131453581},
		{6.671e-09, 12566.151699983, 0.775148887},
	},
	{
		{3.826e-09, 6283.075849991, 5.705257275},
		{3.03e-10, 12566.151699983, 5.407132842},
	},
}

// 地心近似，不包含观测者位置引起的日周项。
func tdbPeriodicSeconds(jd float64) float64 {
	t := (jd - 2451545.0) / 365250.0
	sum := 0.0
	for i := len(tdbPeriodicGroups) - 1; i >= 0; i-- {
		group := 0.0
		for _, term := range tdbPeriodicGroups[i] {
			group += term[0] * math.Sin(term[1]*t+term[2])
		}
		sum = sum*t + group
	}
	return sum + 0.00065e-6*math.Sin(6069.776754*t+4.021194) +
		0.00033e-6*math.Sin(213.299095*t+5.543132) -
		0.00196e-6*math.Sin(6208.294251*t+5.696701) -
		0.00173e-6*math.Sin(74.781599*t+2.435900) +
		0.03638e-6*t*t
}

// TCGMinusTTSeconds 返回 TT 时刻的 TCG−TT（秒）/ TCG−TT in seconds at a TT instant.
func TCGMinusTTSeconds(jd float64) float64 {
	return lgRate / (1 - lgRate) * (jd - barycentricT0JDE) * 86400
}

// TCBMinusTTSeconds 返回 TT 时刻的地心 TCB−TT 近似值（秒）/ geocentric TCB−TT approximation in seconds at a TT instant.
func TCBMinusTTSeconds(jd float64) float64 {
	return (lbRate*(jd-barycentricT0JDE)*86400 + tdbPeriodicSeconds(jd) - tdb0Seconds) / (1 - lbRate)
}

// TDBMinusTTSeconds 返回 TT 时刻的地心 TDB−TT 近似值（秒）/ geocentric TDB−TT approximation in seconds at a TT instant.
func TDBMinusTTSeconds(jd float64) float64 { return tdbPeriodicSeconds(jd) }

// TT2TCG 地球时转地心坐标时 / converts TT to TCG.
func TT2TCG(ttJDE float64) float64 { return ttJDE + TCGMinusTTSeconds(ttJDE)/86400 }

// TCG2TT 地心坐标时转地球时 / converts TCG to TT.
func TCG2TT(tcgJDE float64) float64 { return tcgJDE - lgRate*(tcgJDE-barycentricT0JDE) }

// TT2TCB 地球时转太阳系质心坐标时，采用地心近似 / converts TT to TCB using a geocentric approximation.
func TT2TCB(ttJDE float64) float64 { return ttJDE + TCBMinusTTSeconds(ttJDE)/86400 }

// TCB2TT 是 TT2TCB 的逆 / inverts TT2TCB.
func TCB2TT(tcbJDE float64) float64 {
	// TDB 与 TT 相差不超过 2 ms，拿 TDB 读数当初值可少迭代两轮。
	tt := TCB2TDB(tcbJDE)
	for i := 0; i < 3; i++ {
		tt = tcbJDE - TCBMinusTTSeconds(tt)/86400
	}
	return tt
}

// TT2TDB 地球时转太阳系质心力学时，采用地心近似 / converts TT to TDB using a geocentric approximation.
func TT2TDB(ttJDE float64) float64 { return ttJDE + TDBMinusTTSeconds(ttJDE)/86400 }

// TDB2TT 是 TT2TDB 的逆 / inverts TT2TDB.
func TDB2TT(tdbJDE float64) float64 {
	tt := tdbJDE - TDBMinusTTSeconds(tdbJDE)/86400
	for i := 0; i < 3; i++ {
		tt = tdbJDE - TDBMinusTTSeconds(tt)/86400
	}
	return tt
}

// TCB2TDB 太阳系质心坐标时转质心力学时 / converts TCB to TDB.
func TCB2TDB(tcbJDE float64) float64 {
	return tcbJDE + (tdb0Seconds/86400 - lbRate*(tcbJDE-barycentricT0JDE))
}

// TDB2TCB 太阳系质心力学时转质心坐标时 / converts TDB to TCB.
func TDB2TCB(tdbJDE float64) float64 {
	return tdbJDE + (lbRate*(tdbJDE-barycentricT0JDE)-tdb0Seconds/86400)/(1-lbRate)
}
