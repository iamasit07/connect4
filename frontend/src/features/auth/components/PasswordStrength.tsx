import { Check, X } from 'lucide-react';
import { useMemo } from 'react';

interface PasswordStrengthProps {
  password: string;
}

interface Rule {
  label: string;
  test: (pw: string) => boolean;
}

const rules: Rule[] = [
  { label: 'At least 8 characters', test: (pw) => pw.length >= 8 },
  { label: 'One uppercase letter', test: (pw) => /[A-Z]/.test(pw) },
  { label: 'One lowercase letter', test: (pw) => /[a-z]/.test(pw) },
  { label: 'One digit', test: (pw) => /\d/.test(pw) },
  { label: 'One special character', test: (pw) => /[^a-zA-Z0-9]/.test(pw) },
];

export const PasswordStrength = ({ password }: PasswordStrengthProps) => {
  const results = useMemo(
    () => rules.map((rule) => ({ ...rule, passed: rule.test(password) })),
    [password]
  );

  const passedCount = results.filter((r) => r.passed).length;

  if (!password) return null;

  // Compute a simple strength label
  const strength =
    passedCount <= 2 ? 'Weak' : passedCount <= 4 ? 'Fair' : 'Strong';

  const strengthColor =
    passedCount <= 2
      ? 'bg-destructive'
      : passedCount <= 4
        ? 'bg-yellow-500'
        : 'bg-green-500';

  return (
    <div className="space-y-2 mt-2">
      {/* Strength bar */}
      <div className="flex items-center gap-2">
        <div className="flex-1 h-1.5 bg-muted rounded-full overflow-hidden">
          <div
            className={`h-full ${strengthColor} rounded-full transition-all duration-300`}
            style={{ width: `${(passedCount / rules.length) * 100}%` }}
          />
        </div>
        <span
          className={`text-xs font-medium ${
            passedCount <= 2
              ? 'text-destructive'
              : passedCount <= 4
                ? 'text-yellow-500'
                : 'text-green-500'
          }`}
        >
          {strength}
        </span>
      </div>

      {/* Rule checklist */}
      <ul className="space-y-1">
        {results.map((rule) => (
          <li
            key={rule.label}
            className={`flex items-center gap-1.5 text-xs ${
              rule.passed ? 'text-green-500' : 'text-muted-foreground'
            }`}
          >
            {rule.passed ? (
              <Check className="h-3 w-3" />
            ) : (
              <X className="h-3 w-3" />
            )}
            {rule.label}
          </li>
        ))}
      </ul>
    </div>
  );
};
