---
name: Test Failure Report
about: Report a failing test case
title: "[TEST FAILURE] "
labels: ["bug", "testing"]
assignees: []
---

## Test Failure Details

### Test Information
- **Test Type**: [ ] Unit Test [ ] Acceptance Test [ ] Integration Test
- **Test Name**: 
- **Package**: 
- **Test Command**: 

### Environment
- **OS**: 
- **Go Version**: 
- **Provider Version**: 
- **n8n Version** (for acceptance tests): 
- **Terraform Version** (for acceptance tests): 

### Failure Description
<!-- Describe what the test was trying to do and how it failed -->

### Error Output
```
<!-- Paste the test failure output here -->
```

### Expected Behavior
<!-- What should have happened? -->

### Actual Behavior
<!-- What actually happened? -->

### Reproduction Steps
1. 
2. 
3. 

### Additional Context
<!-- Any additional information that might be relevant -->

### Test Logs
<!-- If available, paste relevant test logs -->

### Configuration
<!-- For acceptance tests, include the Terraform configuration used -->
```hcl

```

### Checklist
- [ ] I have checked that this is not a duplicate issue
- [ ] I have included all relevant environment information
- [ ] I have included the complete error output
- [ ] I have tried reproducing the issue locally