package rum

import (
	"encoding/json"
	"fmt"
	"rum/app/common"
	"sort"
	"strings"
	"time"
)

// ProfileMetric is the top-level metric container saved on the Kit
type ProfileMetric struct {
	Metric map[string]IMetric `json:"metric"` // profile name -> metric
}

// IMetric holds all recorded data for a single profile run
type IMetric struct {
	Profile      map[int]IMetricProfile      `json:"profile"`
	RequestsMade int64                       `json:"requestsMade"`
	Succeed      map[int]IMetricAgentSucceed `json:"succeed"`
	Fail         map[int]IMetricAgentFail    `json:"fail"`
	Misc         map[int]IMetricMisc         `json:"misc"`
	// Buget        map[int]IMetricBudget       `json:"budget"`
}

func NewIMetric() IMetric {
	return IMetric{
		Profile: make(map[int]IMetricProfile),
		Succeed: make(map[int]IMetricAgentSucceed),
		Fail:    make(map[int]IMetricAgentFail),
		Misc:    make(map[int]IMetricMisc),
		// Buget:   make(map[int]IMetricBudget),
	}
}

type IMetricProfile struct {
	Name  string `json:"name"`
	Model string `json:"model"`
}

type IMetricMisc struct {
	RemoveAt     time.Time `json:"removeAt"`
	DeactivateAt time.Time `json:"deactivateAt"`
	ActivateAt   time.Time `json:"activateAt"`
}

type IMetricAgentSucceed struct {
	TimeTaken     time.Duration `json:"timeTaken"`
	ClientRequest string        `json:"clientRequest"`
	AgentReply    string        `json:"agentReply"`
	At            time.Time     `json:"at"`
}

type IMetricAgentFail struct {
	At     time.Time `json:"at"`
	Reason string    `json:"reason"`
}

type IMetricBudget struct {
	Left    float64 `json:"left"`
	Current float64 `json:"current"`
	Set     float64 `json:"set"`
}

func NewProfileMetric() ProfileMetric {
	return ProfileMetric{
		Metric: make(map[string]IMetric),
	}
}

// JSON returns the metric as pretty-printed JSON for readable logging
func (r ProfileMetric) JSON() string {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Sprintf("<json error: %v>", err)
	}
	return string(b)
}

func (r *ProfileMetric) Prompt() string {
	var sb strings.Builder
	if len(r.Metric) == 0 {
		return "<empty metric receipt>"
	}

	profiles := make([]string, 0, len(r.Metric))
	for profile := range r.Metric {
		profiles = append(profiles, profile)
	}
	sort.Strings(profiles)

	sb.WriteString("===== Metric Receipt =====\n")
	for _, profile := range profiles {
		m := r.Metric[profile]
		sb.WriteString(fmt.Sprintf("Profile: %s\n", profile))
		if prof, ok := m.Profile[1]; ok {
			sb.WriteString(fmt.Sprintf("Model: %s\n", prof.Model))
		}
		sb.WriteString(fmt.Sprintf("Requests Made: %d\n", m.RequestsMade))

		if len(m.Profile) > 0 {
			sb.WriteString("Profiles:\n")
			for i := 1; i <= len(m.Profile); i++ {
				if prof, ok := m.Profile[i]; ok {
					sb.WriteString(fmt.Sprintf("  %d. name=%s model=%s\n", i, prof.Name, prof.Model))
				}
			}
		}

		if len(m.Succeed) > 0 {
			sb.WriteString("Successes:\n")
			for i := 1; i <= len(m.Succeed); i++ {
				if succ, ok := m.Succeed[i]; ok {
					t := time.Now().Add(succ.TimeTaken)
					sb.WriteString(fmt.Sprintf("  %d. at=%s duration=%s\n", i, common.GenerateServerTime(succ.At), common.FormatDateForClient(t)))
					sb.WriteString(fmt.Sprintf("     client request: %s\n", succ.ClientRequest))
					sb.WriteString(fmt.Sprintf("     agent reply: %s\n", succ.AgentReply))
				}
			}
		}

		if len(m.Fail) > 0 {
			sb.WriteString("Failures:\n")
			for i := 1; i <= len(m.Fail); i++ {
				if fail, ok := m.Fail[i]; ok {
					sb.WriteString(fmt.Sprintf("  %d. at=%s reason=%s\n", i, common.GenerateServerTime(fail.At), fail.Reason))
				}
			}
		}

		// if len(m.Buget) > 0 {
		// 	sb.WriteString("Budget Events:\n")
		// 	for i := 1; i <= len(m.Buget); i++ {
		// 		if bud, ok := m.Buget[i]; ok {
		// 			sb.WriteString(fmt.Sprintf("  %d. current=%.2f left=%.2f set=%.2f\n", i, bud.Current, bud.Left, bud.Set))
		// 		}
		// 	}
		// }

		if len(m.Misc) > 0 {
			sb.WriteString("Misc Events:\n")
			for i := 1; i <= len(m.Misc); i++ {
				if misc, ok := m.Misc[i]; ok {
					if !misc.RemoveAt.IsZero() {
						sb.WriteString(fmt.Sprintf("  %d. removed at %s\n", i, common.GenerateServerTime(misc.RemoveAt)))
					}
					if !misc.DeactivateAt.IsZero() {
						sb.WriteString(fmt.Sprintf("  %d. deactivated at %s\n", i, common.FormatDateForClient(misc.DeactivateAt)))
					}
					if !misc.ActivateAt.IsZero() {
						sb.WriteString(fmt.Sprintf("  %d. activated at %s\n", i, common.GenerateServerTime(misc.ActivateAt)))
					}
				}
			}
		}

		sb.WriteString("----------------------------\n")
	}
	return sb.String()
}

func (m *IMetric) PCount() int { return len(m.Profile) }
func (m *IMetric) FCount() int { return len(m.Fail) }
func (m *IMetric) SCount() int { return len(m.Succeed) }

// func (m *IMetric) BCount() int { return len(m.Buget) }

func (m *IMetric) MCount() int { return len(m.Misc) }

func (m *IMetric) AddProfile(profile IMetricProfile) {
	m.Profile[m.PCount()+1] = profile
}

func (m *IMetric) AddRequest() {
	m.RequestsMade++
}

// func (m *IMetric) AddBudget(b IMetricBudget) {
// 	m.Buget[m.BCount()+1] = b
// }

func (m *IMetric) AddSucceedReport(resp IMetricAgentSucceed) {
	m.Succeed[m.SCount()+1] = resp
}

func (m *IMetric) AddFailReport(resp IMetricAgentFail) {
	m.Fail[m.FCount()+1] = resp
}

func (m *IMetric) AddRemoveReport(t time.Time) {
	count := m.MCount() + 1
	src := m.Misc[count]
	src.RemoveAt = t
	m.Misc[count] = src // write back — value type
}

func (m *IMetric) AddDeactiveReport(t time.Time) {
	count := m.MCount() + 1
	src := m.Misc[count]
	src.DeactivateAt = t
	m.Misc[count] = src // write back — value type
}

func (m *IMetric) AddActivateReport(t time.Time) {
	count := m.MCount() + 1
	src := m.Misc[count]
	src.ActivateAt = t
	m.Misc[count] = src // write back — value type
}
