package deploy

// import (
//     "context"
//     "fmt"
//     "log"

//     "github.com/containers/podman/v4/pkg/bindings"
//     "github.com/containers/podman/v4/pkg/bindings/containers"
// )

// func StartDeploy() {
//     // Create a context
//     ctx := context.Background()

//     // Connect to the Podman service
//     conn, err := bindings.NewConnection(ctx, "unix:///run/podman/podman.sock") // @todo: add to config?
//     if err != nil {
//         log.Fatalf("Error connecting to Podman service: %s\n", err)
//     }
//     defer conn.Close()

//     // Define the container configuration
//     containerConfig := containers.CreateOptions{
//         Image: "alpine", // @todo: get image from config
//     }

//     // Create the container
//     containerID, err := containers.CreateWithSpec(ctx, conn, &containerConfig)
//     if err != nil {
//         log.Fatalf("Error creating container: %s\n", err)
//     }

//     // Start the container
//     if err := containers.Start(ctx, conn, containerID, nil); err != nil {
//         log.Fatalf("Error starting container: %s\n", err)
//     }

//     // Wait for the container to finish
//     if _, err := containers.Wait(ctx, conn, containerID, "stopped"); err != nil {
//         log.Fatalf("Error waiting for container: %s\n", err)
//     }

//     // Get the container logs
//     logs, err := containers.Logs(ctx, conn, containerID, &containers.LogOptions{
//         Stdout: true,
//     })
//     if err != nil {
//         log.Fatalf("Error getting container logs: %s\n", err)
//     }

//     // Print the container logs
//     fmt.Println(logs)
// }