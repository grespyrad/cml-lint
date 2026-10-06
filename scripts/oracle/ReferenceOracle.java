// Test-only oracle. Production CMLGo/cml-lint require no JVM.
import java.io.File;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import org.eclipse.xtext.validation.CheckMode;
import org.eclipse.xtext.util.CancelIndicator;
import org.eclipse.xtext.diagnostics.Severity;
import org.contextmapper.dsl.standalone.ContextMapperStandaloneSetup;

public final class ReferenceOracle {
  public static void main(String[] args) {
    var api = ContextMapperStandaloneSetup.getStandaloneAPI();
    for (var path : args) {
      try {
        var model = api.loadCML(new File(path));
        var messages = new StringBuilder();
        var resource=model.getXtextResource();
        var issues=resource.getResourceServiceProvider().getResourceValidator().validate(resource,CheckMode.ALL,CancelIndicator.NullImpl);
        boolean valid=true;
        for(var issue:issues) {if(issue.getSeverity()==Severity.ERROR) {valid=false;messages.append(issue.getLineNumber()).append(": ").append(issue.getMessage()).append('\n');}}
        System.out.println(path + "\t" + valid + "\t" + Base64.getEncoder().encodeToString(messages.toString().getBytes(StandardCharsets.UTF_8)));
      } catch (Throwable error) {
        System.out.println(path + "\tcrash\t" + Base64.getEncoder().encodeToString(error.toString().getBytes(StandardCharsets.UTF_8)));
      }
    }
  }
}
