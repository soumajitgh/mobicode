import { Box } from '@/components/ui/box';
import { Text } from '@/components/ui/text';

export function TabPage({ title }: { title: string }) {
  return (
    <Box className="flex-1 items-center justify-center bg-background">
      <Text className="text-foreground" size="xl">
        {title}
      </Text>
    </Box>
  );
}
