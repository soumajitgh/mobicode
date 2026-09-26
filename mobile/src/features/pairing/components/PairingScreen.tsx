import {
  CameraView,
  useCameraPermissions,
  type BarcodeScanningResult,
} from 'expo-camera';
import { Check, ScanLine } from 'lucide-react-native';
import { useEffect, useRef, useState } from 'react';
import {
  ActivityIndicator,
  AppState,
  Linking,
  Pressable,
  StyleSheet,
  Text,
  View,
  useWindowDimensions,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import Animated, {
  useAnimatedStyle,
  useSharedValue,
  withTiming,
} from 'react-native-reanimated';

import { parsePairingPayload } from '@/features/pairing/schemas';
import { theme } from '@/shared/theme';
import { usePairingStore } from '@/store/pairing-store';

export function PairingScreen() {
  const { width, height } = useWindowDimensions();
  const frameSize = Math.min(width - 48, height * 0.34, 336);
  const [permission, requestPermission, getPermission] = useCameraPermissions();
  const requestedPermission = useRef(false);
  const scanLocked = useRef(false);
  const [scanned, setScanned] = useState(false);
  const [success, setSuccess] = useState(false);
  const [scanError, setScanError] = useState<string | null>(null);
  const [cameraError, setCameraError] = useState(false);
  const successProgress = useSharedValue(0);
  const setPayload = usePairingStore((state) => state.setPayload);
  const submit = usePairingStore((state) => state.submit);
  const busy = usePairingStore((state) => state.busy);
  const pairingError = usePairingStore((state) => state.error);

  useEffect(() => {
    if (permission?.status === 'undetermined' && !requestedPermission.current) {
      requestedPermission.current = true;
      void requestPermission();
    }
  }, [permission?.status, requestPermission]);

  useEffect(() => {
    const subscription = AppState.addEventListener('change', (state) => {
      if (state === 'active') void getPermission();
    });
    return () => subscription.remove();
  }, [getPermission]);

  const scanning = permission?.granted && !scanned && !busy && !cameraError;
  const issue =
    scanError ?? pairingError ?? (cameraError ? 'Camera unavailable.' : null);
  const waiting = !permission || scanning || (busy && !success);
  const buttonLabel = success
    ? 'Paired'
    : busy
      ? 'Pairing…'
      : issue
        ? 'Scan again'
        : !permission
          ? 'Preparing camera…'
          : !permission.granted
            ? permission.canAskAgain
              ? 'Allow camera'
              : 'Open settings'
            : 'Looking for QR code';

  useEffect(() => {
    successProgress.value = withTiming(success ? 1 : 0, { duration: 420 });
  }, [success, successProgress]);

  const cornerStyle = useAnimatedStyle(() => ({
    opacity: 1 - successProgress.value,
  }));
  const borderStyle = useAnimatedStyle(() => ({
    opacity: successProgress.value,
  }));
  const checkStyle = useAnimatedStyle(() => ({
    opacity: successProgress.value,
    transform: [{ scale: 0.7 + successProgress.value * 0.3 }],
  }));

  function handleScan({ data }: BarcodeScanningResult) {
    if (scanLocked.current) return;
    scanLocked.current = true;
    setScanned(true);

    try {
      parsePairingPayload(data);
    } catch {
      setScanError('This is not a Mobicode pairing code.');
      return;
    }

    setPayload(data);
    void submit(async () => {
      setSuccess(true);
      await new Promise<void>((resolve) => setTimeout(resolve, 1100));
    }).then(() => {
      if (usePairingStore.getState().error) setSuccess(false);
    });
  }

  function handleButtonPress() {
    if (!permission?.granted) {
      if (permission?.canAskAgain) {
        void requestPermission();
      } else {
        void Linking.openSettings();
      }
      return;
    }

    scanLocked.current = false;
    setScanned(false);
    setSuccess(false);
    setScanError(null);
    setCameraError(false);
    setPayload('');
  }

  return (
    <SafeAreaView style={styles.screen} edges={['top', 'bottom']}>
      <View style={styles.content}>
        <Text style={styles.title}>Pair your phone</Text>

        <View style={[styles.scanner, { width: frameSize, height: frameSize }]}>
          {scanning ? (
            <CameraView
              style={StyleSheet.absoluteFill}
              facing="back"
              barcodeScannerSettings={{ barcodeTypes: ['qr'] }}
              onBarcodeScanned={handleScan}
              onMountError={() => setCameraError(true)}
            />
          ) : success ? (
            <Animated.View style={[styles.check, checkStyle]}>
              <Check size={52} strokeWidth={3} color={successGreen} />
            </Animated.View>
          ) : (
            <ScanLine size={56} strokeWidth={1.3} color={theme.colors.muted} />
          )}
          <Animated.View
            pointerEvents="none"
            style={[styles.corner, styles.topLeft, cornerStyle]}
          />
          <Animated.View
            pointerEvents="none"
            style={[styles.corner, styles.topRight, cornerStyle]}
          />
          <Animated.View
            pointerEvents="none"
            style={[styles.corner, styles.bottomLeft, cornerStyle]}
          />
          <Animated.View
            pointerEvents="none"
            style={[styles.corner, styles.bottomRight, cornerStyle]}
          />
          <Animated.View
            pointerEvents="none"
            style={[styles.successBorder, borderStyle]}
          />
        </View>

        <View style={styles.footer}>
          {(issue || (permission && !permission.granted)) && (
            <Text
              style={[styles.message, issue && styles.errorMessage]}
              accessibilityRole="alert"
            >
              {issue ?? 'Camera access is needed to scan the code.'}
            </Text>
          )}
          <Pressable
            accessibilityRole="button"
            accessibilityLabel={buttonLabel}
            disabled={waiting || success}
            onPress={handleButtonPress}
            style={({ pressed }) => [
              styles.button,
              pressed && styles.buttonPressed,
            ]}
          >
            {waiting && (
              <ActivityIndicator color={theme.colors.background} size="small" />
            )}
            <Text style={styles.buttonText}>{buttonLabel}</Text>
          </Pressable>
        </View>
      </View>
    </SafeAreaView>
  );
}

const successGreen = '#4ade80';

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: theme.colors.background,
    paddingHorizontal: theme.spacing.lg,
  },
  content: {
    flex: 1,
    width: '100%',
    maxWidth: 440,
    alignSelf: 'center',
    paddingTop: theme.spacing.xl + theme.spacing.xs,
    paddingBottom: theme.spacing.lg,
  },
  title: {
    color: theme.colors.text,
    fontSize: theme.typography.title,
    fontWeight: '700',
    letterSpacing: -0.8,
    textAlign: 'center',
  },
  scanner: {
    alignSelf: 'center',
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: theme.spacing.xl * 2,
    backgroundColor: theme.colors.surface,
    borderColor: theme.colors.border,
    borderWidth: 1,
    borderRadius: 28,
    overflow: 'hidden',
  },
  corner: {
    position: 'absolute',
    width: 42,
    height: 42,
    borderColor: theme.colors.primary,
    borderWidth: 5,
  },
  successBorder: {
    position: 'absolute',
    top: 0,
    right: 0,
    bottom: 0,
    left: 0,
    borderColor: successGreen,
    borderWidth: 5,
    borderRadius: 28,
  },
  check: {
    alignItems: 'center',
    justifyContent: 'center',
    width: 88,
    height: 88,
    borderRadius: 44,
    backgroundColor: '#173826',
  },
  topLeft: {
    top: 18,
    left: 18,
    borderTopLeftRadius: 8,
    borderRightWidth: 0,
    borderBottomWidth: 0,
  },
  topRight: {
    top: 18,
    right: 18,
    borderTopRightRadius: 8,
    borderLeftWidth: 0,
    borderBottomWidth: 0,
  },
  bottomLeft: {
    bottom: 18,
    left: 18,
    borderBottomLeftRadius: 8,
    borderRightWidth: 0,
    borderTopWidth: 0,
  },
  bottomRight: {
    bottom: 18,
    right: 18,
    borderBottomRightRadius: 8,
    borderLeftWidth: 0,
    borderTopWidth: 0,
  },
  footer: { marginTop: 'auto', gap: theme.spacing.md },
  message: {
    color: theme.colors.muted,
    fontSize: 14,
    textAlign: 'center',
  },
  errorMessage: { color: theme.colors.danger },
  button: {
    minHeight: 56,
    borderRadius: 18,
    backgroundColor: theme.colors.primary,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: theme.spacing.sm,
  },
  buttonPressed: { opacity: 0.85 },
  buttonText: {
    color: theme.colors.background,
    fontSize: theme.typography.body,
    fontWeight: '600',
  },
});
