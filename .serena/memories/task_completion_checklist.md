# GoFn Task Completion Checklist

## After Making Changes to Source Code

### 1. Code Generation
```fish
# Regenerate code if you added/modified //gofn: directives
go run ./cmd/gofn -src=. -out=.
```

### 2. Code Quality Checks
```fish
# Format code
gofmt -w .

# Static analysis
go vet ./...

# Build verification
go build ./...
```

### 3. Functional Testing
```fish
# If working in example directory
cd example && go run .

# Verify the generated code works as expected
```

### 4. Documentation
- Update README.md if new features were added
- Ensure generated code includes proper headers
- Verify examples are up to date

## Before Committing

### Pre-commit Checklist
- [ ] Code is formatted (`gofmt -w .`)
- [ ] No vet warnings (`go vet ./...`)
- [ ] Project builds successfully (`go build ./...`)
- [ ] Generated files are up to date
- [ ] Examples run without errors
- [ ] Documentation reflects changes

### Git Workflow
```fish
git add .
git commit -m "descriptive message"
git push
```

## Release Checklist (if applicable)
- [ ] Version updated in relevant files
- [ ] README examples are current
- [ ] All generated code is up to date
- [ ] License file is present (MIT)
- [ ] No temporary or debug code remains