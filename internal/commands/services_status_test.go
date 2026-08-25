package commands

import (
	"testing"

	chauftruntime "github.com/siegg/chauffeur/internal/runtime"
)

func TestMergePodmanStatusDeduplicatesStableResources(t *testing.T) {
	byName := make(map[string]chauftruntime.ServiceStatus)
	linked := make(map[string][]string)
	order := []string{}
	status := chauftruntime.ServiceStatus{Name: "chauf-php83-fpm", State: "running", Healthy: true}
	mergePodmanStatus(byName, linked, &order, status, "one")
	mergePodmanStatus(byName, linked, &order, status, "two")
	mergePodmanStatus(byName, linked, &order, status, "one")
	if len(order) != 1 || len(byName) != 1 {
		t.Fatalf("resources were not deduplicated: order=%v statuses=%v", order, byName)
	}
	if len(linked[status.Name]) != 2 || linked[status.Name][0] != "one" || linked[status.Name][1] != "two" {
		t.Fatalf("linked projects = %v; want one,two", linked[status.Name])
	}
}

func TestPodmanStatusOrderPlacesNginxFirst(t *testing.T) {
	statuses := map[string]chauftruntime.ServiceStatus{
		"php":   {Role: "php-fpm"},
		"nginx": {Role: "nginx"},
	}
	got := podmanStatusOrderFirst([]string{"php", "nginx"}, statuses, "nginx")
	if len(got) != 2 || got[0] != "nginx" || got[1] != "php" {
		t.Fatalf("status order = %v; want nginx,php", got)
	}
}
