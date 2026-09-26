import type { ReactNode } from 'react';

import { Pressable } from '@/components/ui/pressable';
import { Text } from '@/components/ui/text';
import { VStack } from '@/components/ui/vstack';
import { theme } from '@/shared/theme';

type BottomTabItemProps = {
  label: string;
  icon: ReactNode;
  active: boolean;
  onPress: () => void;
  onLongPress: () => void;
};

export function BottomTabItem({
  label,
  icon,
  active,
  onPress,
  onLongPress,
}: BottomTabItemProps) {
  return (
    <Pressable
      accessibilityRole="tab"
      accessibilityLabel={label}
      accessibilityState={{ selected: active }}
      onPress={onPress}
      onLongPress={onLongPress}
      className="flex-1"
    >
      <VStack className="min-h-14 items-center justify-center gap-1">
        {icon}
        <Text
          size="xs"
          style={{ color: active ? theme.colors.primary : theme.colors.muted }}
        >
          {label}
        </Text>
      </VStack>
    </Pressable>
  );
}
