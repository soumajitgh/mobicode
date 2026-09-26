import { Tabs } from 'expo-router';
import type { ComponentProps } from 'react';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { HStack } from '@/components/ui/hstack';
import { theme } from '@/shared/theme';
import { BottomTabItem } from './bottom-tab-item';

type BottomTabBarProps = Parameters<
  NonNullable<ComponentProps<typeof Tabs>['tabBar']>
>[0];

export function BottomTabBar({
  state,
  descriptors,
  navigation,
}: BottomTabBarProps) {
  const { bottom } = useSafeAreaInsets();

  return (
    <HStack
      className="border-t border-border bg-card"
      style={{
        paddingTop: theme.spacing.xs,
        paddingBottom: bottom + theme.spacing.xs,
      }}
    >
      {state.routes.map((route, index) => {
        const active = state.index === index;
        const options = descriptors[route.key].options;
        const label =
          typeof options.tabBarLabel === 'string'
            ? options.tabBarLabel
            : (options.title ?? route.name);
        const color = active ? theme.colors.primary : theme.colors.muted;

        return (
          <BottomTabItem
            key={route.key}
            label={label}
            icon={options.tabBarIcon?.({ focused: active, color, size: 22 })}
            active={active}
            onPress={() => {
              const event = navigation.emit({
                type: 'tabPress',
                target: route.key,
                canPreventDefault: true,
              });
              if (!active && !event.defaultPrevented)
                navigation.navigate(route.name, route.params);
            }}
            onLongPress={() =>
              navigation.emit({ type: 'tabLongPress', target: route.key })
            }
          />
        );
      })}
    </HStack>
  );
}
