import React from 'react';
import Svg, { Circle } from 'react-native-svg';

// Diana pequeña — el glifo de "en riesgo de ruleta". Reemplaza al emoji 🎯,
// igual que en la web (regla del proyecto: nada de stickers en la UI).
export default function TargetGlyph({
  size = 14,
  color = '#BC5C42',
}: {
  size?: number;
  color?: string;
}) {
  return (
    <Svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <Circle cx={12} cy={12} r={9.5} stroke={color} strokeWidth={2.4} />
      <Circle cx={12} cy={12} r={5} stroke={color} strokeWidth={2.4} />
      <Circle cx={12} cy={12} r={1.8} fill={color} />
    </Svg>
  );
}
