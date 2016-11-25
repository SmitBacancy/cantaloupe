package services

import(
	"encoding/json"
	// "fmt"
	"time"
)

type IncidentInfo struct {
	Id int
	Component_id int
	Serial_no string
	Title string
	Recorder string
	Component string
	Warranty_timestamp time.Time
	Warranty_till string
	Machine *string
	Description string
	Status string
	Resolved_at *time.Time
}

func AddIncident(data string) {
	sess := SetupDB()
	i := IncidentInfo{}
	err := json.Unmarshal([]byte(data), &i)//converting JSON Object to GO structure ...
	CheckErr(err)
	i.Status = "active"

	_, err2 := sess.InsertInto("incidents").
		Columns("component_id", "title", "recorder", "description", "status").
		Record(i).
		Exec()
	CheckErr(err2)

	_, err3 := sess.Update("components").
		Set("active", false).
		Where("id = ?", i.Component_id).
		Exec()
	CheckErr(err3)

	_, err4 := sess.Update("machine_components").
		Set("deleted_at", "NOW()").
		Where("component_id = ?", i.Component_id).
		Exec()
	CheckErr(err4)
}

func EditIncident(incidentId int, componentId string, recorder string, title string, description string) {
	sess := SetupDB()
	query := sess.Update("incidents")
		if(componentId != "") {
			query.Set("component_id", componentId)
		}
		query.Set("title", title).
		Set("description", description).
		Set("recorder", recorder).
		Where("id = ?", incidentId).
		Exec()
}

func DeleteIncident(incidentId int) {
	sess := SetupDB()

	componentId, err := sess.Select("component_id").
		From("incidents").
		Where("id = ?", incidentId).
		ReturnInt64()
	CheckErr(err)

	_, err1 := sess.Update("components").
		Set("active", true).
		Where("id = ?", componentId).
		Exec()
	CheckErr(err1)

	_, err2 := sess.DeleteFrom("incidents").
		Where("id = ?", incidentId).
		Exec()
	CheckErr(err2)
}

func DisplayIncidents() []byte {
	sess := SetupDB()
	incidentInfo := []IncidentInfo{}

	query := sess.Select("i.id, i.title, i.description, i.status, i.recorder, components.name AS Component, components.serial_no,components.warranty_till AS Warranty_timestamp, machines.name AS Machine").
		From("incidents i").
		LeftJoin("components", "i.component_id = components.id").
		LeftJoin("machine_components", "i.component_id = machine_components.component_id").
		LeftJoin("machines", "machines.id = machine_components.component_id")

		query.LoadStruct(&incidentInfo)

	//extract only date from timestamp========
	for i := 0; i < len(incidentInfo); i++ {
		t := incidentInfo[i].Warranty_timestamp
		incidentInfo[i].Warranty_till = t.Format("2006-01-02")
	}
	//================================

	b, err := json.Marshal(incidentInfo)
	CheckErr(err)
	return b
}

func DisplayIncident(incidentId int) []byte {
	sess := SetupDB()
	incidentInfo := IncidentInfo{}
	sess.Select("id, component_id, title, description, status, resolved_at").
		From("incidents").
		Where("id = ?", incidentId).
		LoadStruct(&incidentInfo)
	b, err := json.Marshal(incidentInfo)
	CheckErr(err)
	return b
}

type IncidentUpdate struct {
	Id int
	Resolved_by string
	Created_at time.Time
	Resolved_Date string
	ComponentId int64
	IncidentId int
	Description string
}

func IncidentUpdates(incidentId int, resolvedBy string, description string) {
	sess := SetupDB()

	//select component id on which incident happen ....==========
	id, err1 := sess.Select("component_id").
		From("incidents").
		Where("id = ?", incidentId).
		ReturnInt64()
	CheckErr(err1)
	//===========================================================

 //insert record into incident_updates table ...===============
	var incident IncidentUpdate
	incident.ComponentId = id
	incident.IncidentId = incidentId
	incident.Resolved_by = resolvedBy
	incident.Description = description

	_, err := sess.InsertInto("incident_update").
		Columns("incident_id", "component_id", "description", "resolved_by").
		Record(incident).
		Exec()
	CheckErr(err)
	//===========================================================



	// //Add replaces components after resolved ....================
	// components := DisplayAllComponents{}
	// components.Active = true
	// err2 := sess.Select("name, invoice_id, warranty_till, description, active").
	// 	From("components").
	// 	Where("id = ? ", id).
	// 	LoadStruct(&components)
	// CheckErr(err2)

	// _, err3 := sess.InsertInto("components").
	// 	Columns("name", "invoice_id", "warranty_till", "description", "active").
	// 	Record(components).
	// 	Exec()
	// CheckErr(err3)
	// //===========================================================

	// //===Add resolved date in incidents table after resolved component
	// _, err5 := sess.Update("incidents").
	// 	Set("resolved_at", "NOW()").
	// 	Set("status", "Resloved").
	// 	Where("id = ?", incidentId).
	// 	Exec()
	// CheckErr(err5)
	// //===========================================================
}

type IncidentInformation struct {
	Status string
	Recorder string
	Machine string
	Component string
	ComponentId int
	Description string

	IncidentUpdates[] IncidentUpdate
}

func IncidentInformations(incident_id int) []byte{
	sess := SetupDB()
	m := IncidentInformation {}
	err2 := sess.Select("incidents.status, incidents.recorder, incidents.component_id, components.name as Component, incidents.description").
		From("incidents").
		LeftJoin("components", "components.id = incidents.component_id").
		Where("incidents.id = ? ", incident_id).
		LoadStruct(&m)
	CheckErr(err2)

	p := []IncidentUpdate {}
		err3 := sess.Select("id, description, resolved_by, created_at").
		From("incident_update").
		Where("incident_id = ? ", incident_id).
		LoadStruct(&p)
	CheckErr(err3)

	for i := 0; i < len(p); i++ {
		t := p[i].Created_at
		p[i].Resolved_Date = t.Format("2006-01-02")
	}

	m.IncidentUpdates = p
	b, err := json.Marshal(m)
	CheckErr(err)
	return b
}
