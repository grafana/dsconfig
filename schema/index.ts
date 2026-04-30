// Schema types
export type {
    DatasourceConfigSchema,
    ConfigField,
    FieldItemSchema,
    Expression,
    ValueType,
    SemanticType,
    FieldKind,
    Lifecycle,
    TargetLocation,
    UIComponent,
    UIWidth,
    FieldUI,
    FieldValidationRule,
    PatternValidationRule,
    RangeValidationRule,
    LengthValidationRule,
    ItemCountValidationRule,
    AllowedValuesValidationRule,
    CustomValidationRule,
    BaseValidationRule,
    FieldOverride,
    StorageMapping,
    DirectMapping,
    IndexedPairMapping,
    ComputedMapping,
    MappingField,
    FieldOption,
    ConfigGroup,
    RelationshipType,
    FieldRelationship,
    SecureFieldState,
    FormState,
} from "./schema"

// Validation
export {
    validateSchema,
    validateField,
    validateValidationRule,
    validateStorageMapping,
} from "./validate"
export type { ValidationError } from "./validate"

// Runtime guards
export {
    isValueType,
    isSemanticType,
    isFieldKind,
    isLifecycle,
    isTargetLocation,
    isUIComponent,
    isUIWidth,
    isRelationshipType,
    isStorageMapping,
    isValidationRule,
    isValidOptionValue,
} from "./guards"

// Runtime loader and validator
export {
    loadAndValidate,
    newDatasourceConfig,
} from "./runtime"
export type {
    LoadMode,
    DatasourceConfig,
    ConfigError,
    SecureState,
    ValueSource,
    FieldValue,
    LoadResult,
} from "./runtime"
