package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type OverviewService struct {
	dao *mysql.DAO
}

func NewOverviewService(dao *mysql.DAO) *OverviewService {
	return &OverviewService{
		dao: dao,
	}
}

const (
	overviewDaysPerWeek   = 7
	overviewHistoryWeeks  = 8
	overviewWeekCount     = overviewHistoryWeeks + 1
	overviewBeijingOffset = 8 * time.Hour
)

var overviewBeijing = time.FixedZone("Asia/Shanghai", int(overviewBeijingOffset/time.Second))

func overviewWeeks(now time.Time) []dto.OverviewWeek {
	local := now.In(overviewBeijing)
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, overviewBeijing)
	// 将周日视为一周最后一天，统一回退到北京时间周一零点。
	daysSinceMonday := (int(day.Weekday()) - int(time.Monday) + overviewDaysPerWeek) % overviewDaysPerWeek
	monday := day.AddDate(0, 0, -daysSinceMonday)
	weeks := make([]dto.OverviewWeek, overviewWeekCount)
	for i := range weeks {
		start := monday.AddDate(0, 0, (i-overviewHistoryWeeks)*overviewDaysPerWeek)
		weeks[i] = dto.OverviewWeek{
			Start:     start,
			End:       start.AddDate(0, 0, overviewDaysPerWeek-1),
			IsCurrent: i == overviewHistoryWeeks,
		}
	}
	return weeks
}

// 按发布记录的构建版本号判断新旧，避免版本名字符串排序导致 0.0.10 排在 0.0.9 之前。
func overviewVersionUsage(reports []mysql.OverviewReportedVersion, versions []models.AppVersion) (dto.OverviewVersionUsage, *models.AppVersion) {
	const (
		groupLatest = iota
		groupPrevious
		groupOlder
		groupUnreported
		groupOther
	)

	sort.Slice(versions, func(i, j int) bool {
		if versions[i].VersionCode == versions[j].VersionCode {
			return versions[i].ID > versions[j].ID
		}
		return versions[i].VersionCode > versions[j].VersionCode
	})
	var latest, previous *models.AppVersion
	for i := range versions {
		if versions[i].Status == models.AppVersionStatusPublished {
			latest = &versions[i]
			break
		}
	}
	if latest != nil {
		// 上一发布版本允许已归档，但构建版本号必须低于当前已发布版本。
		for i := range versions {
			if versions[i].VersionCode < latest.VersionCode {
				previous = &versions[i]
				break
			}
		}
	}
	codes := make(map[string]int, len(versions))
	for _, version := range versions {
		name := strings.TrimSpace(version.VersionName)
		if _, exists := codes[name]; !exists {
			codes[name] = version.VersionCode
		}
	}
	groups := []dto.OverviewVersionGroup{
		groupLatest:     {Key: "latest"},
		groupPrevious:   {Key: "previous"},
		groupOlder:      {Key: "older"},
		groupUnreported: {Key: "unreported"},
		groupOther:      {Key: "other"},
	}
	if latest != nil {
		groups[groupLatest].Version = strings.TrimSpace(latest.VersionName)
	}
	if previous != nil {
		groups[groupPrevious].Version = strings.TrimSpace(previous.VersionName)
	}
	var total int64
	for _, report := range reports {
		name := strings.TrimSpace(report.Version)
		code, known := codes[name]
		// 无法匹配发布记录的版本归入“其他”，不根据版本名推测它是否更早。
		index := groupOther
		switch {
		case name == "":
			index = groupUnreported
		case latest != nil && name == groups[groupLatest].Version:
			index = groupLatest
		case previous != nil && name == groups[groupPrevious].Version:
			index = groupPrevious
		case previous != nil && known && code < previous.VersionCode:
			index = groupOlder
		}
		groups[index].Count += report.Count
		total += report.Count
	}
	latestCount := groups[groupLatest].Count
	if total == 0 {
		groups = make([]dto.OverviewVersionGroup, 0)
	} else {
		for i := range groups {
			groups[i].Percentage = float64(groups[i].Count) / float64(total) * 100
		}
		if groups[groupOther].Count == 0 {
			groups = groups[:groupOther]
		}
	}
	return dto.OverviewVersionUsage{
		Total:       total,
		Groups:      groups,
		LatestCount: latestCount,
	}, latest
}

func (o *OverviewService) GetAdminOverview() (*dto.AdminOverview, error) {
	// 只取一次当前时间，让各模块共用同一统计截止时间。
	return o.getAdminOverviewAt(time.Now())
}

func (o *OverviewService) getAdminOverviewAt(now time.Time) (*dto.AdminOverview, error) {
	result := &dto.AdminOverview{
		GeneratedAt: now.In(overviewBeijing),
	}
	err := o.dao.OverviewReadTransaction(func(tx *gorm.DB) error {
		var err error
		result.Statistics, err = o.dao.GetOverviewStatistics(tx, now)
		if err != nil {
			return err
		}
		result.WeeklyApplications, err = o.getOverviewWeeklyApplications(tx, now)
		if err != nil {
			return err
		}
		result.RecentRecords, err = o.dao.GetOverviewRecentRecords(tx, now)
		if err != nil {
			return err
		}
		result.VersionUsage, result.LatestAndroidVersion, err = o.getOverviewClientVersions(tx, now)
		if err != nil {
			return err
		}
		return nil
	})
	// 任一模块查询失败，整体返回错误，避免把不完整结果展示为真实统计。
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o *OverviewService) getOverviewWeeklyApplications(tx *gorm.DB, now time.Time) (dto.OverviewWeeklyApplications, error) {
	result := dto.OverviewWeeklyApplications{
		Weeks: overviewWeeks(now),
	}
	weeks := result.Weeks
	counts, err := o.dao.GetOverviewWeekCounts(tx, weeks[0].Start, weeks[len(weeks)-1].Start.AddDate(0, 0, overviewDaysPerWeek), now)
	if err != nil {
		return dto.OverviewWeeklyApplications{}, err
	}
	// 数据库只返回有申请的周；预先生成的九个节点保留零值，避免折线图缺少空周。
	for _, count := range counts {
		if count.WeekIndex >= 0 && count.WeekIndex < len(weeks) {
			weeks[count.WeekIndex].Count = count.Count
			result.Total += count.Count
		}
	}
	return result, nil
}

// 版本使用情况与版本卡片共用发布历史，避免重复查询及“最新版本”口径不一致。
func (o *OverviewService) getOverviewClientVersions(tx *gorm.DB, now time.Time) (dto.OverviewVersionUsage, *dto.OverviewVersion, error) {
	reports, err := o.dao.GetOverviewReportedVersions(tx)
	if err != nil {
		return dto.OverviewVersionUsage{}, nil, err
	}
	versions, err := o.dao.GetOverviewAndroidVersions(tx, now)
	if err != nil {
		return dto.OverviewVersionUsage{}, nil, err
	}
	usage, latest := overviewVersionUsage(reports, versions)
	if latest == nil {
		return usage, nil, nil
	}
	version, err := adminAppVersion(latest)
	if err != nil {
		return dto.OverviewVersionUsage{}, nil, err
	}
	return usage, &dto.OverviewVersion{
		VersionName:  version.VersionName,
		VersionCode:  version.VersionCode,
		PublishedAt:  version.PublishedAt,
		ReleaseNotes: version.ReleaseNotes,
	}, nil
}
