# Contributing to tiny-pushgateway

Thank you for your interest in contributing to tiny-pushgateway!

## How to Contribute

### Reporting Issues

- Check if the issue already exists
- Provide a clear description of the problem
- Include steps to reproduce
- Specify your environment (OS, Go version, etc.)

### Submitting Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Add/update tests as needed
5. Run tests: `go test -v -race ./...`
6. Commit with a clear message
7. Push to your fork
8. Open a Pull Request

### Code Guidelines

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep it simple - this project values simplicity and minimal dependencies
- Add tests for new functionality
- Update documentation as needed

### Testing

```bash
# Run tests
go test -v ./...

# Run with race detector
go test -v -race ./...

# Build and test Docker image
docker build -t tiny-pushgateway:test .
docker run -p 9091:9091 tiny-pushgateway:test
```

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

