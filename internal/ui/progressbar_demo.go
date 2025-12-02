// +build ignore

package main

import (
	"time"

	"gothicforge3/internal/ui"
)

func main() {
	// Demo 1: Basic progress bar
	println("Demo 1: Basic Progress Bar")
	bar := ui.NewProgressBar(100, "Downloading files...")
	for i := 0; i < 100; i += 5 {
		bar.Add(5)
		time.Sleep(50 * time.Millisecond)
	}
	bar.Finish()
	println()

	// Demo 2: Progress bar with description updates
	println("Demo 2: Progress Bar with Dynamic Descriptions")
	bar2 := ui.NewProgressBar(100, "Phase 1: Initializing...")
	bar2.Add(25)
	time.Sleep(300 * time.Millisecond)

	bar2.UpdateDescription("Phase 2: Processing data...")
	bar2.Add(25)
	time.Sleep(300 * time.Millisecond)

	bar2.UpdateDescription("Phase 3: Finalizing...")
	bar2.Add(25)
	time.Sleep(300 * time.Millisecond)

	bar2.UpdateDescription("Phase 4: Complete!")
	bar2.Add(25)
	bar2.Finish()
	println()

	// Demo 3: Deployment simulation
	println("Demo 3: Deployment Simulation")
	deployBar := ui.NewProgressBar(100, "Starting deployment...")
	
	deployBar.Set(10)
	time.Sleep(200 * time.Millisecond)
	deployBar.UpdateDescription("Provisioning database...")
	
	deployBar.Set(30)
	time.Sleep(400 * time.Millisecond)
	deployBar.UpdateDescription("Provisioning cache...")
	
	deployBar.Set(50)
	time.Sleep(400 * time.Millisecond)
	deployBar.UpdateDescription("Running migrations...")
	
	deployBar.Set(70)
	time.Sleep(300 * time.Millisecond)
	deployBar.UpdateDescription("Deploying application...")
	
	deployBar.Set(90)
	time.Sleep(300 * time.Millisecond)
	deployBar.UpdateDescription("Running health checks...")
	
	deployBar.Set(100)
	deployBar.UpdateDescription("Deployment complete!")
	deployBar.Finish()
	
	println("\n🎉 All demos completed!")
}
