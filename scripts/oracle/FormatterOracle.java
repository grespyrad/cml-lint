// Тестовый вызов настоящего Xtext IFormatter2; production не требует JVM.
import java.io.File;
import java.nio.file.Files;
import java.nio.file.Path;
import org.contextmapper.dsl.standalone.ContextMapperStandaloneSetup;
import org.eclipse.xtext.formatting2.FormatterRequest;
import org.eclipse.xtext.formatting2.IFormatter2;
import org.eclipse.xtext.formatting2.regionaccess.TextRegionAccessBuilder;
public final class FormatterOracle {
  public static void main(String[] args) throws Exception {
    var api = ContextMapperStandaloneSetup.getStandaloneAPI();
    for (int i = 0; i < args.length; i += 2) {
      var path = args[i];
      var resource = api.loadCML(new File(path)).getXtextResource();
      var provider = resource.getResourceServiceProvider();
      var access = provider.get(TextRegionAccessBuilder.class).forNodeModel(resource).create();
      var request = new FormatterRequest().setTextRegionAccess(access);
      var edits = provider.get(IFormatter2.class).format(request);
      Files.writeString(Path.of(args[i + 1]), access.getRewriter().renderToString(edits));
    }
  }
}
