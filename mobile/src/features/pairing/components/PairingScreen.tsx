import {
  CameraView,
  useCameraPermissions,
  type BarcodeScanningResult,
} from 'expo-camera';
import { ScanLine } from 'lucide-react-native';
import { useEffect, useRef, useState } from 'react';
import {
  ActivityIndicator,
  Animated,
  AppState,
  Easing,
  Linking,
  Pressable,
  StyleSheet,
  Text,
  View,
  useWindowDimensions,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { parsePairingPayload } from '@/features/pairing/schemas';
import { theme } from '@/shared/theme';
import { usePairingStore } from '@/store/pairing-store';

export function PairingScreen() {
  const { width, height } = useWindowDimensions();
  const frameSize = Math.min(width - 48, height * 0.34, 336);
  const [permission, requestPermission, getPermission] = useCameraPermissions();
  const requestedPermission = useRef(false);
  const scanLocked = useRef(false);
  const [scanProgress] = useState(() => new Animated.Value(0));
  const [scanned, setScanned] = useState(false);
  const [scanError, setScanError] = useState<string | null>(null);
  const [cameraError, setCameraError] = useState(false);
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
    scanError ?? pairingError ?? (cameraError ? 'camera unavailable.' : null);
  const waiting = !permission || scanning || busy;
  const buttonLabel = busy
    ? 'pairing…'
    : issue
      ? 'scan again'
      : !permission
        ? 'preparing camera…'
        : !permission.granted
          ? permission.canAskAgain
            ? 'allow camera'
            : 'open settings'
          : 'looking for qr code';

  useEffect(() => {
    if (!scanning) return;
    scanProgress.setValue(0);
    const animation = Animated.loop(
      Animated.sequence([
        Animated.timing(scanProgress, {
          toValue: 1,
          duration: 1900,
          easing: Easing.inOut(Easing.ease),
          isInteraction: false,
          useNativeDriver: true,
        }),
        Animated.timing(scanProgress, {
          toValue: 0,
          duration: 1900,
          easing: Easing.inOut(Easing.ease),
          isInteraction: false,
          useNativeDriver: true,
        }),
      ]),
    );
    animation.start();
    return () => animation.stop();
  }, [scanning, scanProgress]);

  function handleScan({ data }: BarcodeScanningResult) {
    if (scanLocked.current) return;
    scanLocked.current = true;
    setScanned(true);

    try {
      parsePairingPayload(data);
    } catch {
      setScanError('this is not a mobicode pairing code.');
      return;
    }

    setPayload(data);
    void submit();
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
    setScanError(null);
    setCameraError(false);
    setPayload('');
  }

  return (
    <SafeAreaView style={styles.screen} edges={['top', 'bottom']}>
      <View style={styles.content}>
        <Text style={styles.title}>pair your phone</Text>

        <View style={[styles.scanner, { width: frameSize, height: frameSize }]}>
          {scanning ? (
            <CameraView
              style={StyleSheet.absoluteFill}
              facing="back"
              barcodeScannerSettings={{ barcodeTypes: ['qr'] }}
              onBarcodeScanned={handleScan}
              onMountError={() => setCameraError(true)}
            />
          ) : (
            <ScanLine size={56} strokeWidth={1.3} color={theme.colors.muted} />
          )}
          {scanning && (
            <Animated.View
              pointerEvents="none"
              style={[
                styles.scanLine,
                {
                  transform: [
                    {
                      translateY: scanProgress.interpolate({
                        inputRange: [0, 1],
                        outputRange: [-frameSize / 2 + 52, frameSize / 2 - 52],
                      }),
                    },
                  ],
                },
              ]}
            />
          )}
          <View pointerEvents="none" style={[styles.corner, styles.topLeft]} />
          <View pointerEvents="none" style={[styles.corner, styles.topRight]} />
          <View
            pointerEvents="none"
            style={[styles.corner, styles.bottomLeft]}
          />
          <View
            pointerEvents="none"
            style={[styles.corner, styles.bottomRight]}
          />
        </View>

        <View style={styles.footer}>
          {(issue || (permission && !permission.granted)) && (
            <Text
              style={[styles.message, issue && styles.errorMessage]}
              accessibilityRole="alert"
            >
              {issue ?? 'camera access is needed to scan the code.'}
            </Text>
          )}
          <Pressable
            accessibilityRole="button"
            accessibilityLabel={buttonLabel}
            disabled={waiting}
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

const scannerAccent = '#82e8b2';

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
    fontFamily: 'monospace',
    fontVariant: ['small-caps'],
    fontSize: 26,
    fontWeight: '700',
    letterSpacing: 2,
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
  scanLine: {
    position: 'absolute',
    top: '50%',
    left: '18%',
    width: '64%',
    height: 2,
    borderRadius: 1,
    backgroundColor: scannerAccent,
    shadowColor: scannerAccent,
    shadowOpacity: 0.8,
    shadowRadius: 10,
    elevation: 4,
  },
  corner: {
    position: 'absolute',
    width: 42,
    height: 42,
    borderColor: scannerAccent,
    borderWidth: 3,
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
    fontFamily: 'monospace',
    fontVariant: ['small-caps'],
    fontSize: 12,
    letterSpacing: 0.8,
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
    fontFamily: 'monospace',
    fontVariant: ['small-caps'],
    fontSize: 13,
    fontWeight: '700',
    letterSpacing: 1.6,
  },
});
