package main

import (
  "github.com/gizak/termui"
  "github.com/jszwedko/go-circleci"
  "github.com/sirupsen/logrus"
)

func doCircleCI() ([]*termui.Table, error) {
  // Check that the CircleCI API token is not empty
  if len(circleciToken) <= 0 {
    logrus.Warn("CircleCI API token cannot be empty")
    logrus.Info("skipping CircleCI data")
    return nil, nil
  }

  // Check that the CircleCI owners are not empty.
  if len(circleciOwners) <= 0 {
    logrus.Warn("CircleCI owners cannot be empty")
    logrus.Info("skipping CircleCI data")
    return nil, nil
  }

  tables := []*termui.Table{}

  // Initialize the CircleCI client
  circleciClient := &circleci.Client{Token: circleciToken}
  projects, _ := circleciClient.ListProjects()

  // Bucketize project per owner
  projectsForOwner := make(map[string][]*circleci.Project)
  for _, project := range projects {
    projectsForOwner[project.Username] = append(projectsForOwner[project.Username], project)
  }

  for _, owner := range circleciOwners {
    // Initialize the table.
    table := termui.NewTable()
    rows := [][]string{
      {"repo", "branch", "state", "finished at"},
    }
    redrows := []int{}
    greenrows := []int{}
    otherrows := []int{}

    // TODO: sort the project list for nicer display 
    for _, project := range projectsForOwner[owner] {
      builds, _ := circleciClient.ListRecentBuildsForProject(project.Username, project.Reponame, "master", "", 1, 0)
      for _, build := range builds {
        if showAllBuilds || build.Status != "success" {
          rows = append(rows, []string{build.Reponame, build.Branch, build.Status, build.StopTime.String()})

          if build.Status == "failed" {
            redrows = append(redrows, len(rows)-1)
          } else if build.Status == "fixed" {
            greenrows = append(greenrows, len(rows)-1)
          } else if build.Status != "success" {
            otherrows = append(otherrows, len(rows)-1)
          }
        }
      }
    }

    // Set the rows.
    table.Rows = rows

    // Set the default colors and settings
    table.FgColor = termui.ColorWhite
    table.BgColor = termui.ColorDefault
    table.TextAlign = termui.AlignLeft
    table.Border = true
    table.Separator = true
    table.Block.BorderLabel = "CircleCI build for " + owner
    table.Analysis()
    table.SetSize()

    // Set the color to red for the red rows
    for _, br := range redrows {
      table.FgColors[br] = termui.ColorRed
    }
    // Set the color to green for the green rows
    for _, br := range greenrows {
      table.FgColors[br] = termui.ColorGreen
    }
    // Set the color to yellow for the other rows
    for _, br := range otherrows {
      table.FgColors[br] = termui.ColorYellow
    }

    tables = append(tables, table)
  }
  return tables, nil
}

func circleciWidget(body *termui.Grid) {
  if body == nil {
    body = termui.Body
  }

  circleci, err := doCircleCI()
  if err != nil {
    logrus.Fatal(err)
  }
  if circleci != nil {
    columns := []*termui.Row{}
    for _, t := range circleci {
      columns = append(columns, termui.NewCol(int(12/len(circleci)), 0, t))
    }
    body.AddRows(termui.NewRow(columns...))

    // Calculate the layout.
    body.Align()
    // Render the termui body.
    termui.Render(body)
  }
}
