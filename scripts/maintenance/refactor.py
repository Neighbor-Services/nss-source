import re
import os

# 1. Provider Remote Data Source
file_ds = 'e:/ns/frontend/nsapp/lib/features/provider/data/datasource/remote/provider_remote_datasource.dart'
with open(file_ds, 'r', encoding='utf-8') as f:
    c = f.read()
if 'verifyAppointmentCode(' not in c:
    c = c.replace('Future<List<AppointmentData>> getAppointments();', 'Future<List<AppointmentData>> getAppointments();\n  Future<bool> verifyAppointmentCode(String appointmentId, String code);')
    with open(file_ds, 'w', encoding='utf-8') as f: f.write(c)

# 2. Provider Remote Data Source Impl
file_ds_impl = 'e:/ns/frontend/nsapp/lib/features/provider/data/datasource/remote/provider_remote_datasource_impl.dart'
with open(file_ds_impl, 'r', encoding='utf-8') as f:
    c = f.read()
if 'verifyAppointmentCode(' not in c:
    impl = """
  @override
  Future<bool> verifyAppointmentCode(String appointmentId, String code) async {
    final token = await Helpers.getString("token");
    final response = await dio.post(
      "$baseUrl/interactions/appointments/$appointmentId/verify-code/",
      data: {'code': code},
      options: Options(headers: dioHeaders(token)),
    );
    if (response.statusCode == 200) {
      return true;
    } else {
      throw ServerException();
    }
  }
"""
    c = c.replace('class ProviderRemoteDataSourceImpl implements ProviderRemoteDataSource {', 'class ProviderRemoteDataSourceImpl implements ProviderRemoteDataSource {' + impl)
    with open(file_ds_impl, 'w', encoding='utf-8') as f: f.write(c)

# 3. Provider Repository
file_repo = 'e:/ns/frontend/nsapp/lib/features/provider/domain/repository/provider_repository.dart'
with open(file_repo, 'r', encoding='utf-8') as f:
    c = f.read()
if 'verifyAppointmentCode(' not in c:
    c = c.replace('Future<Either<Failure, List<AppointmentData>>> getAppointments();', 'Future<Either<Failure, List<AppointmentData>>> getAppointments();\n  Future<Either<Failure, bool>> verifyAppointmentCode(String appointmentId, String code);')
    with open(file_repo, 'w', encoding='utf-8') as f: f.write(c)

# 4. Provider Repository Impl
file_repo_impl = 'e:/ns/frontend/nsapp/lib/features/provider/data/repository/provider_repository_impl.dart'
with open(file_repo_impl, 'r', encoding='utf-8') as f:
    c = f.read()
if 'verifyAppointmentCode(' not in c:
    impl = """
  @override
  Future<Either<Failure, bool>> verifyAppointmentCode(String appointmentId, String code) async {
    if (await networkInfo.isConnected) {
      try {
        final result = await remoteDataSource.verifyAppointmentCode(appointmentId, code);
        return Right(result);
      } catch (e) {
        return Left(ServerFailure());
      }
    } else {
      return Left(NetworkFailure());
    }
  }
"""
    c = c.replace('class ProviderRepositoryImpl implements ProviderRepository {', 'class ProviderRepositoryImpl implements ProviderRepository {' + impl)
    with open(file_repo_impl, 'w', encoding='utf-8') as f: f.write(c)

# 5. Use Case (NEW FILE)
uc_path = 'e:/ns/frontend/nsapp/lib/features/provider/domain/usecase/verify_appointment_code_use_case.dart'
uc_content = """import 'package:dartz/dartz.dart';
import 'package:nsapp/core/models/failure.dart';
import 'package:nsapp/features/provider/domain/repository/provider_repository.dart';

class VerifyAppointmentCodeUseCase {
  final ProviderRepository repository;
  VerifyAppointmentCodeUseCase(this.repository);

  Future<Either<Failure, bool>> call(String appointmentId, String code) async {
    return await repository.verifyAppointmentCode(appointmentId, code);
  }
}
"""
with open(uc_path, 'w', encoding='utf-8') as f: f.write(uc_content)

# 6. Provider Event
file_event = 'e:/ns/frontend/nsapp/lib/features/provider/presentation/bloc/provider_event.dart'
with open(file_event, 'r', encoding='utf-8') as f:
    c = f.read()
if 'VerifyAppointmentCodeEvent' not in c:
    evt = """
class VerifyAppointmentCodeEvent extends ProviderEvent {
  final String appointmentId;
  final String code;
  VerifyAppointmentCodeEvent({required this.appointmentId, required this.code});
}
"""
    c = c + evt
    with open(file_event, 'w', encoding='utf-8') as f: f.write(c)

# 7. Provider State
file_state = 'e:/ns/frontend/nsapp/lib/features/provider/presentation/bloc/provider_state.dart'
with open(file_state, 'r', encoding='utf-8') as f:
    c = f.read()
if 'VerifyAppointmentCodeState' not in c:
    st = """
class VerifyAppointmentCodeLoadingState extends ProviderState {}
class SuccessVerifyAppointmentCodeState extends ProviderState {}
class FailureVerifyAppointmentCodeState extends ProviderState {}
"""
    c = c + st
    with open(file_state, 'w', encoding='utf-8') as f: f.write(c)

# 8. Provider BLOC
file_bloc = 'e:/ns/frontend/nsapp/lib/features/provider/presentation/bloc/provider_bloc.dart'
with open(file_bloc, 'r', encoding='utf-8') as f:
    c = f.read()
if 'VerifyAppointmentCodeEvent' not in c:
    c = c.replace('import \'package:nsapp/features/provider/domain/usecase/update_appointment_use_case.dart\';', 'import \'package:nsapp/features/provider/domain/usecase/update_appointment_use_case.dart\';\nimport \'package:nsapp/features/provider/domain/usecase/verify_appointment_code_use_case.dart\';')
    c = c.replace('final GetRequestDetailUseCase getRequestDetailUseCase;', 'final GetRequestDetailUseCase getRequestDetailUseCase;\n  final VerifyAppointmentCodeUseCase verifyAppointmentCodeUseCase;')
    c = c.replace('this.getRequestDetailUseCase,\n  ) : super(ProviderInitial()) {', 'this.getRequestDetailUseCase,\n    this.verifyAppointmentCodeUseCase,\n  ) : super(ProviderInitial()) {')
    
    bloc_handler = """
    on<VerifyAppointmentCodeEvent>((event, emit) async {
      emit(VerifyAppointmentCodeLoadingState());
      final results = await verifyAppointmentCodeUseCase(event.appointmentId, event.code);
      results.fold(
        (l) => emit(FailureVerifyAppointmentCodeState()),
        (r) => emit(SuccessVerifyAppointmentCodeState()),
      );
    });
"""
    c = c.replace('on<ProviderEvent>((event, emit) {});', 'on<ProviderEvent>((event, emit) {});' + bloc_handler)
    with open(file_bloc, 'w', encoding='utf-8') as f: f.write(c)

# 9. Injection Container
file_inj = 'e:/ns/frontend/nsapp/lib/core/di/injection_container.dart'
with open(file_inj, 'r', encoding='utf-8') as f:
    c = f.read()
if 'VerifyAppointmentCodeUseCase' not in c:
    c = c.replace('import \'package:nsapp/features/provider/domain/usecase/update_appointment_use_case.dart\';', 'import \'package:nsapp/features/provider/domain/usecase/update_appointment_use_case.dart\';\nimport \'package:nsapp/features/provider/domain/usecase/verify_appointment_code_use_case.dart\';')
    c = c.replace('  sl.registerLazySingleton(() => UpdateProviderAppointmentUseCase(sl()));', '  sl.registerLazySingleton(() => UpdateProviderAppointmentUseCase(sl()));\n  sl.registerLazySingleton(() => VerifyAppointmentCodeUseCase(sl()));')
    # Use re to match the ProviderBloc constructor args to append sl() safely
    import re
    c = re.sub(r'(sl\(\),\s*sl\(\),\s*\)\);)', r'sl(),\n        \1', c) 
    with open(file_inj, 'w', encoding='utf-8') as f: f.write(c)

# 10. Provider UI Refactor
file_ui = 'e:/ns/frontend/nsapp/lib/features/shared/presentation/widget/appointment_detail_bottom_sheet.dart'
with open(file_ui, 'r', encoding='utf-8') as f:
    c = f.read()

verify_code_ui_old = '''  Future<void> _verifyCode() async {
    final appt = widget.data.appointment;
    if (appt == null || _codeController.text.trim().isEmpty) return;
    setState(() => _isVerifying = true);
    try {
      final token = await Helpers.getString("token");
      final response = await dio.post(
        "$baseUrl/interactions/appointments/${appt.id}/verify-code/",
        data: {'code': _codeController.text.trim()},
        options: Options(headers: dioHeaders(token)),
      );
      if (response.statusCode == 200) {
        setState(() {
           appt.status = 'IN_PROGRESS';
        });
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text("Code verified successfully!")),
          );
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.response?.data['error'] ?? "Failed to verify code")),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text("Error verifying code")),
        );
      }
    } finally {
      if (mounted) setState(() => _isVerifying = false);
    }
  }'''

verify_code_ui_new = '''  void _verifyCode() {
    final appt = widget.data.appointment;
    if (appt == null || _codeController.text.trim().isEmpty) return;
    context.read<ProviderBloc>().add(
      VerifyAppointmentCodeEvent(appointmentId: appt.id!, code: _codeController.text.trim()),
    );
  }'''
c = c.replace(verify_code_ui_old, verify_code_ui_new)

if 'BlocListener<ProviderBloc, ProviderState>' not in c:
    c = c.replace('return Padding(', '''return BlocListener<ProviderBloc, ProviderState>(
      listener: (context, state) {
        if (state is VerifyAppointmentCodeLoadingState) {
          setState(() => _isVerifying = true);
        } else if (state is SuccessVerifyAppointmentCodeState) {
          setState(() {
            _isVerifying = false;
            widget.data.appointment?.status = 'IN_PROGRESS';
          });
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text("Code verified successfully!")),
          );
        } else if (state is FailureVerifyAppointmentCodeState) {
          setState(() => _isVerifying = false);
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text("Failed to verify code. Please try again.")),
          );
        }
      },
      child: Padding(''')
    c = c.replace('''    );
  }
}''', '''    ),
    );
  }
}''')

with open(file_ui, 'w', encoding='utf-8') as f:
    f.write(c)

print("SUCCESSFULLY REFACTORED")
