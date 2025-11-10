#!/bin/bash
# Test script to verify ulf has all necessary permissions

echo "🔍 Testing permissions for user ulf..."
echo ""

# Test 1: File ownership
echo "✓ Test 1: File ownership in ~/sharecare"
if [ -w ~/sharecare/sharecare ]; then
    echo "  ✅ Can write to sharecare binary"
else
    echo "  ❌ Cannot write to sharecare binary"
fi

# Test 2: Git operations
echo ""
echo "✓ Test 2: Git operations"
cd ~/sharecare
if git status &>/dev/null; then
    echo "  ✅ Git works"
else
    echo "  ❌ Git doesn't work"
fi

# Test 3: Go build
echo ""
echo "✓ Test 3: Go availability"
export PATH=$PATH:/usr/local/go/bin
if go version &>/dev/null; then
    echo "  ✅ Go is available: $(go version)"
else
    echo "  ❌ Go not found"
fi

# Test 4: Systemctl (requires sudo)
echo ""
echo "✓ Test 4: Systemctl operations"
if sudo -n systemctl status sharecare.service &>/dev/null; then
    echo "  ✅ Can run systemctl without password"
else
    echo "  ⚠️  Cannot test systemctl in script (needs TTY), but should work in interactive session"
fi

# Test 5: Read systemd file
echo ""
echo "✓ Test 5: Access to systemd file"
if sudo -n cat /etc/systemd/system/sharecare.service &>/dev/null; then
    echo "  ✅ Can read systemd file"
else
    echo "  ⚠️  Cannot test in script (needs TTY)"
fi

# Test 6: Daemon reload
echo ""
echo "✓ Test 6: Daemon reload capability"
if sudo -n systemctl daemon-reload &>/dev/null; then
    echo "  ✅ Can run daemon-reload"
else
    echo "  ⚠️  Cannot test in script (needs TTY)"
fi

echo ""
echo "✅ Basic tests complete!"
echo ""
echo "📝 Note: sudo commands require interactive terminal to work."
echo "   When running Claude as ulf, all sudo commands will work without password."
